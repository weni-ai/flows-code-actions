def Run(engine):
    # Exemplo de uso de secrets
    # Os secrets são carregados automaticamente do projeto associado ao código
    
    # Método 1: Usando get() com valor default
    api_key = engine.secrets.get("API_KEY", "default-key")
    
    # Método 2: Acesso direto como dicionário (lança KeyError se não existir)
    # db_password = engine.secrets["DB_PASSWORD"]
    
    # Método 3: Verificar se existe antes de usar
    if engine.secrets.has("OPENAI_API_KEY"):
        openai_key = engine.secrets.get("OPENAI_API_KEY")
        engine.log.info(f"OpenAI API Key loaded: {openai_key[:8]}...")
    
    # Método 4: Usando operador 'in'
    if "DATABASE_URL" in engine.secrets:
        engine.log.info("Database URL is configured")
    
    # Listar todos os nomes de secrets disponíveis (não os valores)
    available_secrets = engine.secrets.keys()
    engine.log.info(f"Available secrets: {available_secrets}")
    
    # Resultado
    engine.result.set({
        "message": "Action executed successfully!",
        "secrets_loaded": len(available_secrets),
        "api_key_preview": api_key[:8] + "..." if len(api_key) > 8 else api_key
    }, content_type="json")
