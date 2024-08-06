import pymongo
from bson.objectid import ObjectId

import argparse
import action

client = pymongo.MongoClient("mongodb://localhost:27017")
db = client["code-actions"]

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
    def set(self, value):
        self._result = str(value)
        self.save()
    def save(self):
        result = db["coderun"].update_one({"_id": ObjectId(self._runId)}, {"$set": {"result": self._result}})
        if result.modified_count == 1:
            print("result saved!")
        

class Engine:
    def __init__(self, params=Params({}), body="", result=Result("")):
        self.params = params
        self.body = body
        self.result = result


def main():
    parser = argparse.ArgumentParser(description='Parse key-value arguments')
    parser.add_argument('-a', '--arg', action='append', help='Add an argument in the form of key=value')
    parser.add_argument('-b', '--body', type=str, help='Body content')
    parser.add_argument('-r', '--run', type=str, help='run id')

    args = parser.parse_args()

    params_dict = {}
    for arg in args.arg:
        key, value = arg.split('===')
        key = key.strip()
        value = value.strip()
        params_dict[key] = value

    for key, value in params_dict.items():
        print(f"{key}={value}")
        
    body = args.body.strip()
    run_id = args.run.strip()
    
    engine = Engine(params=Params(params_dict), body=body, result=Result(db=db, runId = run_id))
    action.Run(engine)

if __name__ == "__main__":
    main()
