import requests
import json

def Run(engine):
    for chave, valor in engine.params.items():
        print(f"Chave: {chave}, Valor: {valor}")
        
    user_id = engine.params.get("user_id") #engine.params.get("<param_key>")
    usuario = consultar_usuario(user_id)

    if usuario:
        print(usuario)
    else:
        print("Usuário não encontrado.")
    
    print("\n")
    print(engine.body)
    # engine.result.set(json.dumps(usuario))
    engine.log.debug(usuario)
    engine.result.set(usuario)

def consultar_usuario(user_id):

  url = f"https://jsonplaceholder.typicode.com/users/{user_id}"
  response = requests.get(url)

  if response.status_code == 200:
    return response.text
  else:
    return None

