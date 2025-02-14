import os
import pymongo
from bson.objectid import ObjectId
import datetime
import json

import argparse
import action

dburl = os.environ.get("FLOWS_CODE_ACTIONS_MONGO_DB_URI", "mongodb://localhost:27017")
dbname = os.environ.get("FLOWS_CODE_ACTIONS_MONGO_DB_NAME", "code-actions")
client = pymongo.MongoClient(dburl)
db = client[dbname]

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
    def __init__(self, result=None, db=None, runId=None):
        self._result = result
        self._db = db
        self._runId = runId
        self._extra = None
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
        result = db["coderun"].update_one({"_id": ObjectId(self._runId)}, {"$set": {"result": self._result, "extra": self._extra}})
        if result.modified_count == 1:
            print("result saved!")

class Log:
    def __init__(self, db=None, runId=None, codeId=None):
        self._db = db
        self._runId = runId
        self._codeId = codeId

    def _create(self, logtype="", content=""):
        now = datetime.datetime.now()
        log = {"run_id": ObjectId(self._runId), "code_id": ObjectId(self._codeId), "type": logtype, "content": content, "created_at": now, "updated_at": now}
        db["codelog"].insert_one(log)

    def debug(self, content=""):
        self._create(logtype="debug", content=content)
    def info(self, content=""):
        self._create(logtype="info", content=content)
    def error(self, content=""):
        self._create(logtype="error", content=content)

class Engine:
    def __init__(self, params=Params({}), body="", result=Result(""), log=Log()):
        self.params = params
        self.body = body
        self.result = result
        self.log = log


def main():
    parser = argparse.ArgumentParser(description='Parse key-value arguments')
    parser.add_argument('-a', '--arg', type=str, help='Add an argument in the form of key===value')
    parser.add_argument('-b', '--body', type=str, help='Body content')
    parser.add_argument('-r', '--run', type=str, help='run id')
    parser.add_argument('-c', '--codeid', type=str, help='code id')

    args = parser.parse_args()
    
    params_dict = {}
    if args.arg != None:
        params_dict = json.loads(args.arg)
    
    body = ""
    if args.body != None:
        body = args.body.strip()

    run_id = args.run.strip()
    code_id = args.codeid.strip()
    
    engine = Engine(
        params=Params(params_dict), 
        body=body, 
        result=Result(db=db, runId=run_id),
        log=Log(db=db, runId=run_id, codeId=code_id)
    )
    action.Run(engine)

if __name__ == "__main__":
    main()
