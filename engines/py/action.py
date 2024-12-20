import requests
import json

def Run(engine):
    for chave, valor in engine.params.items(): # iterates over query string parameters
        print(f"key: {chave}, value: {valor}")
        
    user_id = engine.params.get("user_id") # to get user_id query string param: engine.params.get("<param_key>")
    user = get_user(user_id)

    if user:
        engine.result.set(user, conten_type="json")
        return # return to finish action execution
    
    engine.result.set({"message": "user not found."}, content_type="json")

def get_user(user_id):

  url = f"https://jsonplaceholder.typicode.com/users/{user_id}"
  response = requests.get(url)

  if response.status_code == 200:
    return response.text
  else:
    return None
