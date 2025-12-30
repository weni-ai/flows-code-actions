-- Queries SQL úteis para verificação após migração MongoDB → PostgreSQL

-- ============================================
-- VERIFICAÇÃO DE CONTAGENS
-- ============================================

-- Contar registros em cada tabela
SELECT 'codes' as table_name, COUNT(*) as total FROM codes
UNION ALL
SELECT 'codelibs', COUNT(*) FROM codelibs
UNION ALL
SELECT 'coderuns', COUNT(*) FROM coderuns
UNION ALL
SELECT 'projects', COUNT(*) FROM projects
UNION ALL
SELECT 'user_permissions', COUNT(*) FROM user_permissions
ORDER BY table_name;

-- ============================================
-- VERIFICAÇÃO DE INTEGRIDADE
-- ============================================

-- Verificar registros sem mongo_object_id (não deveria ter nenhum)
SELECT 'codes_sem_mongo_id' as issue, COUNT(*) as count FROM codes WHERE mongo_object_id IS NULL
UNION ALL
SELECT 'codelibs_sem_mongo_id', COUNT(*) FROM codelibs WHERE mongo_object_id IS NULL
UNION ALL
SELECT 'coderuns_sem_mongo_id', COUNT(*) FROM coderuns WHERE mongo_object_id IS NULL;

-- Verificar registros com datas inválidas (futuras)
SELECT 'codes_futuro' as issue, COUNT(*) as count 
FROM codes WHERE created_at > NOW()
UNION ALL
SELECT 'codelibs_futuro', COUNT(*) 
FROM codelibs WHERE created_at > NOW()
UNION ALL
SELECT 'coderuns_futuro', COUNT(*) 
FROM coderuns WHERE created_at > NOW();

-- Verificar coderuns órfãos (code_id nulo ou inválido)
SELECT 
    COUNT(*) as coderuns_orfaos,
    COUNT(*) FILTER (WHERE code_id IS NULL) as sem_code_id,
    COUNT(*) FILTER (WHERE code_id IS NOT NULL) as com_code_id_invalido
FROM coderuns cr
WHERE NOT EXISTS (
    SELECT 1 FROM codes c WHERE c.id = cr.code_id
);

-- ============================================
-- ESTATÍSTICAS POR TABELA
-- ============================================

-- Estatísticas de codes
SELECT 
    'codes' as tabela,
    COUNT(*) as total,
    COUNT(DISTINCT project_uuid) as projetos_unicos,
    COUNT(*) FILTER (WHERE type = 'flow') as flows,
    COUNT(*) FILTER (WHERE type = 'endpoint') as endpoints,
    COUNT(*) FILTER (WHERE language = 'python') as python,
    COUNT(*) FILTER (WHERE language = 'go') as go,
    COUNT(*) FILTER (WHERE language = 'javascript') as javascript,
    MIN(created_at) as primeiro_registro,
    MAX(created_at) as ultimo_registro
FROM codes;

-- Estatísticas de codelibs
SELECT 
    'codelibs' as tabela,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE language = 'python') as python,
    MIN(created_at) as primeiro_registro,
    MAX(created_at) as ultimo_registro
FROM codelibs;

-- Estatísticas de coderuns
SELECT 
    'coderuns' as tabela,
    COUNT(*) as total,
    COUNT(DISTINCT code_id) as codes_unicos,
    COUNT(*) FILTER (WHERE status = 'queued') as queued,
    COUNT(*) FILTER (WHERE status = 'started') as started,
    COUNT(*) FILTER (WHERE status = 'completed') as completed,
    COUNT(*) FILTER (WHERE status = 'failed') as failed,
    MIN(created_at) as primeiro_registro,
    MAX(created_at) as ultimo_registro
FROM coderuns;

-- ============================================
-- VERIFICAÇÃO DE RELACIONAMENTOS
-- ============================================

-- Top 10 codes com mais execuções
SELECT 
    c.id,
    c.name,
    c.type,
    c.language,
    COUNT(cr.id) as total_execucoes,
    COUNT(*) FILTER (WHERE cr.status = 'completed') as sucesso,
    COUNT(*) FILTER (WHERE cr.status = 'failed') as falhas
FROM codes c
LEFT JOIN coderuns cr ON cr.code_id = c.id
GROUP BY c.id, c.name, c.type, c.language
ORDER BY total_execucoes DESC
LIMIT 10;

-- Códigos sem execuções
SELECT 
    COUNT(*) as codes_sem_execucoes
FROM codes c
WHERE NOT EXISTS (
    SELECT 1 FROM coderuns cr WHERE cr.code_id = c.id
);

-- ============================================
-- VERIFICAÇÃO DE CAMPOS JSON
-- ============================================

-- Verificar estrutura de campos JSONB em coderuns
SELECT 
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE extra IS NOT NULL AND extra != '{}') as com_extra,
    COUNT(*) FILTER (WHERE params IS NOT NULL AND params != '{}') as com_params,
    COUNT(*) FILTER (WHERE headers IS NOT NULL AND headers != '{}') as com_headers
FROM coderuns;

-- Amostra de campos JSON (primeiros 5)
SELECT 
    id,
    status,
    extra,
    params,
    headers
FROM coderuns
WHERE extra IS NOT NULL OR params IS NOT NULL OR headers IS NOT NULL
LIMIT 5;

-- ============================================
-- ANÁLISE TEMPORAL
-- ============================================

-- Distribuição de registros por mês (codes)
SELECT 
    DATE_TRUNC('month', created_at) as mes,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE type = 'flow') as flows,
    COUNT(*) FILTER (WHERE type = 'endpoint') as endpoints
FROM codes
GROUP BY DATE_TRUNC('month', created_at)
ORDER BY mes DESC
LIMIT 12;

-- Distribuição de execuções por dia (últimos 30 dias)
SELECT 
    DATE_TRUNC('day', created_at) as dia,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE status = 'completed') as sucesso,
    COUNT(*) FILTER (WHERE status = 'failed') as falhas,
    ROUND(AVG(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) * 100, 2) as taxa_sucesso
FROM coderuns
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE_TRUNC('day', created_at)
ORDER BY dia DESC;

-- ============================================
-- VERIFICAÇÃO DE DUPLICATAS
-- ============================================

-- Verificar mongo_object_id duplicados (não deveria ter)
SELECT 
    'codes' as tabela,
    mongo_object_id,
    COUNT(*) as duplicatas
FROM codes
GROUP BY mongo_object_id
HAVING COUNT(*) > 1
UNION ALL
SELECT 
    'codelibs',
    mongo_object_id,
    COUNT(*)
FROM codelibs
GROUP BY mongo_object_id
HAVING COUNT(*) > 1
UNION ALL
SELECT 
    'coderuns',
    mongo_object_id,
    COUNT(*)
FROM coderuns
GROUP BY mongo_object_id
HAVING COUNT(*) > 1;

-- ============================================
-- ANÁLISE DE PERFORMANCE
-- ============================================

-- Tamanho das tabelas
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    pg_total_relation_size(schemaname||'.'||tablename) AS size_bytes
FROM pg_tables
WHERE schemaname = 'public'
    AND tablename IN ('codes', 'codelibs', 'coderuns', 'projects', 'user_permissions')
ORDER BY size_bytes DESC;

-- Verificar uso de índices
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
    AND tablename IN ('codes', 'codelibs', 'coderuns', 'projects', 'user_permissions')
ORDER BY tablename, indexname;

-- ============================================
-- COMPARAÇÃO ANTES/DEPOIS
-- ============================================

-- Use estas queries antes e depois da migração para comparar

-- Exemplo: Total de registros por collection/tabela
-- MongoDB:
-- db.code.countDocuments()
-- db.codelib.countDocuments()
-- db.coderun.countDocuments()

-- PostgreSQL:
SELECT COUNT(*) FROM codes;
SELECT COUNT(*) FROM codelibs;
SELECT COUNT(*) FROM coderuns;

-- ============================================
-- LIMPEZA (USE COM CUIDADO!)
-- ============================================

-- Remover registros de teste (descomente se necessário)
-- DELETE FROM coderuns WHERE created_at < '2020-01-01';
-- DELETE FROM codes WHERE created_at < '2020-01-01' AND NOT EXISTS (
--     SELECT 1 FROM coderuns WHERE code_id = codes.id
-- );

-- Vacuum para recuperar espaço (execute após limpeza)
-- VACUUM ANALYZE codes;
-- VACUUM ANALYZE codelibs;
-- VACUUM ANALYZE coderuns;

-- ============================================
-- QUERIES DE EXEMPLO PARA APLICAÇÃO
-- ============================================

-- Buscar code por mongo_object_id (como era no MongoDB)
SELECT * FROM codes WHERE mongo_object_id = 'seu_mongo_id_aqui';

-- Listar codes de um projeto com execuções recentes
SELECT 
    c.*,
    (SELECT COUNT(*) FROM coderuns cr WHERE cr.code_id = c.id) as total_runs,
    (SELECT MAX(created_at) FROM coderuns cr WHERE cr.code_id = c.id) as last_run
FROM codes c
WHERE c.project_uuid = 'seu_project_uuid_aqui'
ORDER BY c.created_at DESC;

-- Buscar execuções de um código específico
SELECT 
    cr.*
FROM coderuns cr
JOIN codes c ON c.id = cr.code_id
WHERE c.mongo_object_id = 'seu_mongo_id_aqui'
ORDER BY cr.created_at DESC
LIMIT 100;

