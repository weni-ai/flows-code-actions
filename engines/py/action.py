def Run(engine):
    # Teste logs em S3
    engine.log.info("Teste de log INFO no S3")
    engine.log.debug("Teste de log DEBUG no S3") 
    engine.log.error("Teste de log ERROR no S3")
    
    # Log com conteúdo longo
    long_content = "Este é um teste de log longo. " * 100
    engine.log.info(f"Log longo: {long_content}")
    
    # Resultado
    engine.result.set({
        "message": "Logs enviados para S3 com sucesso!",
        "timestamp": "2024-12-15T22:00:00Z"
    }, content_type="json")
