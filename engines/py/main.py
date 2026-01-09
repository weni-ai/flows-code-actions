import os
import datetime
import json
import boto3
import uuid
from urllib.parse import urljoin
import requests
import psycopg2
import psycopg2.extras

import argparse
import action

# PostgreSQL configuration
pg_database = os.environ.get("FLOWS_CODE_ACTIONS_DB_NAME", "code-actions")
pg_uri = os.environ.get("FLOWS_CODE_ACTIONS_DB_URI", "postgres://test:test@localhost:5432/codeactions?sslmode=disable")

# Initialize PostgreSQL connection
pg_conn = None
try:
    pg_conn = psycopg2.connect(pg_uri)
    print(f"PostgreSQL connected - database: {pg_database}")
except Exception as e:
    print(f"Failed to connect to PostgreSQL: {e}")
    pg_conn = None

# S3 configuration
s3_enabled = os.environ.get("FLOWS_CODE_ACTIONS_S3_ENABLED", "false").lower() == "true"
s3_endpoint = os.environ.get("FLOWS_CODE_ACTIONS_S3_ENDPOINT") or None  # Treat empty string as None
s3_region = os.environ.get("FLOWS_CODE_ACTIONS_S3_REGION", "us-east-1")
s3_bucket = os.environ.get("FLOWS_CODE_ACTIONS_S3_BUCKET_NAME")
s3_prefix = os.environ.get("FLOWS_CODE_ACTIONS_S3_PREFIX", "codeactions")
s3_access_key = os.environ.get("FLOWS_CODE_ACTIONS_S3_ACCESS_KEY_ID", "test")
s3_secret_key = os.environ.get("FLOWS_CODE_ACTIONS_S3_SECRET_ACCESS_KEY", "test")

# Initialize S3 client if enabled
s3_client = None
if s3_enabled and s3_bucket:
    try:
        session_config = {
            'region_name': s3_region
        }
        
        if s3_access_key and s3_secret_key:
            session_config.update({
                'aws_access_key_id': s3_access_key,
                'aws_secret_access_key': s3_secret_key
            })
        
        if s3_endpoint:
            # For LocalStack or custom S3 endpoints
            s3_client = boto3.client(
                's3',
                endpoint_url=s3_endpoint,
                **session_config
            )
        else:
            # For AWS S3
            s3_client = boto3.client('s3', **session_config)
            
        print(f"S3 codelog enabled - bucket: {s3_bucket}, prefix: {s3_prefix}")
    except Exception as e:
        print(f"Failed to initialize S3 client: {e}")
        s3_enabled = False

def generate_s3_key(run_id, code_id, log_id, timestamp):
    """Generate S3 key following the same pattern as Go implementation:
    {prefix}/logs/{year}/{month}/{day}/{run_id}/{code_id}/{log_id}.json
    """
    year = timestamp.year
    month = timestamp.month
    day = timestamp.day
    
    key_parts = [
        s3_prefix,
        "logs",
        f"{year:04d}",
        f"{month:02d}",
        f"{day:02d}",
        run_id,
        code_id,
        f"{log_id}.json"
    ]
    
    return "/".join(key_parts)

def create_codelog_s3(run_id, code_id, log_type, content):
    """Create a codelog entry in S3 following the Go implementation structure"""
    if not s3_enabled or not s3_client:
        return None
    
    try:
        # Generate new UUID v4 for the log
        log_id = str(uuid.uuid4())
        
        # Create timestamps
        now = datetime.datetime.now(datetime.timezone.utc)
        
        # Create log structure matching Go CodeLog struct
        log_data = {
            "id": log_id,
            "run_id": run_id,
            "code_id": code_id,
            "type": log_type,
            "content": content[:8000],  # Match the 8000 char limit from Go service
            "created_at": now.isoformat(),
            "updated_at": now.isoformat()
        }
        
        # Generate S3 key
        key = generate_s3_key(run_id, code_id, log_id, now)
        
        # Convert to JSON
        json_data = json.dumps(log_data)
        
        # Upload to S3 with same metadata as Go implementation
        s3_client.put_object(
            Bucket=s3_bucket,
            Key=key,
            Body=json_data.encode('utf-8'),
            ContentType='application/json',
            Metadata={
                'run-id': run_id,
                'code-id': code_id,
                'log-type': log_type,
                'created-at': now.isoformat()
            }
        )
        
        print(f"Log saved to S3: {key}")
        return log_id
        
    except Exception as e:
        print(f"Failed to save log to S3: {e}")
        # Don't fail completely, just log the error
        return None

class Params:
    def __init__(self, params={}):
        self._params = params
    def get(self, key):
        if key in self._params:
            return self._params[key]
        return None
    def items(self):
        return self._params.items()
    
class Result:
    def __init__(self, result=None, runId=None, pg_conn=None):
        self._result = result
        self._runId = runId
        self._extra = None
        self._pg_conn = pg_conn
        
    def set(self, value="", status_code=200, content_type="text"):
        if isinstance(value, str):
            self._result = value
        else:
            try:
                self._result = json.dumps(value)
            except:
                self._result = str(value)

        self._extra = {"status_code": status_code, "content_type": content_type}
        self.save()
        
    def save(self):
        """Save result to PostgreSQL"""
        if not self._pg_conn:
            print("Error: PostgreSQL connection not available")
            return
            
        try:
            cursor = self._pg_conn.cursor()
            
            # Convert extra dict to JSON string for PostgreSQL JSONB
            extra_json = json.dumps(self._extra) if self._extra else None
            
            # Update coderuns table in PostgreSQL
            # Using id::text to handle UUID comparison properly
            cursor.execute(
                """
                UPDATE coderuns 
                SET result = %s, 
                    extra = %s::jsonb, 
                    updated_at = NOW()
                WHERE id::text = %s
                """,
                (self._result, extra_json, self._runId)
            )
            
            self._pg_conn.commit()
            
            if cursor.rowcount > 0:
                print(f"Result saved to PostgreSQL! (run_id: {self._runId})")
            else:
                print(f"Warning: No rows updated in PostgreSQL for run_id: {self._runId}")
            
            cursor.close()
        except Exception as e:
            print(f"Failed to save result to PostgreSQL: {e}")
            # Rollback on error
            if self._pg_conn:
                self._pg_conn.rollback()

class Log:
    def __init__(self, runId=None, codeId=None):
        self._runId = runId
        self._codeId = codeId
        self._log_queue = []  # Queue to store logs until flush

    def _create(self, logtype="", content=""):
        """Queue a log entry to be processed later"""
        log_entry = {
            "type": logtype,
            "content": str(content),
            "timestamp": datetime.datetime.now()
        }
        self._log_queue.append(log_entry)
        return f"queued_log_{len(self._log_queue)}"  # Return a placeholder ID
    
    def _process_log(self, logtype, content, timestamp):
        """Create a log entry using S3"""
        if not s3_enabled or not s3_client:
            print("Warning: S3 is not enabled or configured, logs will not be saved")
            return None
            
        # Use S3 implementation
        log_id = create_codelog_s3(self._runId, self._codeId, logtype, content)
        if log_id:
            return log_id
        
        print(f"Failed to save log to S3")
        return None
    
    def flush_logs(self):
        """Process all queued logs"""
        processed_count = 0
        failed_count = 0
        
        print(f"Processing {len(self._log_queue)} queued logs...")
        
        for log_entry in self._log_queue:
            try:
                log_id = self._process_log(
                    log_entry["type"], 
                    log_entry["content"], 
                    log_entry["timestamp"]
                )
                if log_id:
                    processed_count += 1
                else:
                    failed_count += 1
            except Exception as e:
                print(f"Failed to process log: {e}")
                failed_count += 1
        
        print(f"Log processing complete: {processed_count} successful, {failed_count} failed")
        self._log_queue.clear()  # Clear the queue after processing

    def debug(self, content=""):
        """Create a debug log entry"""
        return self._create(logtype="debug", content=content)
        
    def info(self, content=""):
        """Create an info log entry"""
        return self._create(logtype="info", content=content)
        
    def error(self, content=""):
        """Create an error log entry"""  
        return self._create(logtype="error", content=content)

class Header:
    def __init__(self, header={}):
        self._header = header
    def get(self, key):
        if key in self._header:
            return self._header[key][0]
        return ""
    def items(self):
        return self._header.items()

class Request:
    def __init__(self, params=Params({}), body="", log=Log(), header=Header({})):
        self.header = header
        self.params = params
        self.body = body
        self.log = log

class Engine:
    def __init__(self, params=Params({}), body="", result=Result(""), log=Log(), header=Header({}), request=Request()):
        self.params = params
        self.body = body
        self.result = result
        self.log = log
        self.request = request
        self.header = header




def main():
    parser = argparse.ArgumentParser(description='Parse key-value arguments')
    parser.add_argument('-a', '--arg', type=str, help='Add an argument in the form of key===value')
    parser.add_argument('-H', '--header', type=str, help='Header content')
    parser.add_argument('-b', '--body', type=str, help='Body content')
    parser.add_argument('-r', '--run', type=str, help='run id')
    parser.add_argument('-c', '--codeid', type=str, help='code id')

    args = parser.parse_args()
    
    header_dict = {}
    if args.header != None:
        header_dict = json.loads(args.header)
    
    params_dict = {}
    if args.arg != None:
        params_dict = json.loads(args.arg)
    
    body = ""
    if args.body != None:
        body = args.body.strip()

    run_id = args.run.strip()
    code_id = args.codeid.strip()

    header = Header(header_dict)
    params = Params(params_dict)
    result = Result(runId=run_id, pg_conn=pg_conn)
    log = Log(runId=run_id, codeId=code_id)
    request = Request(params=params, body=body, header=header)

    engine = Engine(
        params=params, 
        body=body, 
        result=result,
        log=log,
        header=header,
        request=request,
    )
    try:
        action.Run(engine)
    except Exception as e:
        print(f"Error during action execution: {e}")
        # Log the error to the queue
        log.error(f"Action execution failed: {str(e)}")
    
    # Process all queued logs at the end
    print("Flushing logs...")
    log.flush_logs()
    
    # Close connections
    if pg_conn:
        pg_conn.close()
        print("PostgreSQL connection closed")

if __name__ == "__main__":
    main()
