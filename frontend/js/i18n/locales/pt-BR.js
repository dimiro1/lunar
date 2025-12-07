/**
 * @fileoverview Portuguese Brazil (pt-BR) translations for Lunar Dashboard.
 */

export default {
  // Common/shared strings
  common: {
    loading: "Carregando...",
    save: "Salvar",
    cancel: "Cancelar",
    delete: "Excluir",
    create: "Criar",
    search: "Buscar",
    close: "Fechar",
    enabled: "Ativado",
    disabled: "Desativado",
    noDescription: "Sem descrição",
    na: "N/D",
    invalidDate: "Data inválida",
    back: "Voltar",
    saveChanges: "Salvar alterações",
    functionNotFound: "Função não encontrada",
    status: {
      success: "SUCESSO",
      error: "ERRO",
      timeout: "TIMEOUT",
    },
  },

  // Navigation
  nav: {
    dashboard: "Painel",
    logout: "Sair",
    search: "Buscar",
  },

  // Login page
  login: {
    title: "Lunar",
    subtitle: "Digite sua chave de API para continuar",
    apiKeyLabel: "Chave de API",
    apiKeyPlaceholder: "Digite sua chave de API",
    loginButton: "Entrar",
    loggingIn: "Entrando...",
    invalidKey: "Chave de API inválida",
    footer:
      "Verifique os logs do servidor para obter sua chave de API se este for o seu primeiro acesso.",
  },

  // Functions list
  functions: {
    title: "Funções",
    subtitle: "Gerencie suas funções serverless",
    newFunction: "Nova Função",
    allFunctions: "Todas as Funções",
    totalCount: "{{count}} funções no total",
    emptyState: "Nenhuma função ainda. Crie sua primeira função para começar.",
    loadingFunctions: "Carregando funções...",
    loadingFunction: "Carregando função...",
    columns: {
      name: "Nome",
      description: "Descrição",
      status: "Status",
      version: "Versão",
    },
  },

  // Function tabs
  tabs: {
    code: "Código",
    versions: "Versões",
    executions: "Execuções",
    settings: "Configurações",
    test: "Teste",
  },

  // Function creation
  create: {
    title: "Criar Nova Função",
    subtitle: "Inicialize uma nova função serverless usando Lua.",
    functionName: "Nome da Função",
    functionNamePlaceholder: "ex: webhook-pagamento",
    starterTemplate: "Template Inicial",
    createButton: "Criar Função",
    failedToCreate: "Falha ao criar função",
  },

  // Templates
  templates: {
    http: {
      name: "Template HTTP",
      description: "Manipule requisições HTTP com lógica personalizada",
    },
    api: {
      name: "REST API",
      description: "Construa endpoints de API RESTful",
    },
    aiChat: {
      name: "Chatbot IA",
      description: "API de chat com OpenAI ou Anthropic",
    },
    email: {
      name: "Enviar Email",
      description: "Envie emails via API Resend",
    },
    blank: {
      name: "Em branco",
      description: "Comece com template vazio",
    },
  },

  // Settings page
  settings: {
    generalConfig: "Configuração Geral",
    functionName: "Nome da Função",
    functionDescriprion: "descrição da função",
    description: "Descrição",
    logRetention: "Período de Retenção de Logs",
    retentionHelp:
      "Execuções mais antigas que isso serão excluídas automaticamente",
    envVars: "Variáveis de Ambiente",
    variablesCount: "{{count}} variáveis",
    network: "Endpoint",
    invocationUrl: "URL de Invocação",
    supportedMethods: "Métodos Suportados",
    functionStatus: "Status da Função",
    enableFunction: "Habilitar Função",
    disableWarning:
      "Desabilitar irá parar todas as requisições recebidas por esta função.",
    dangerZone: "Zona de Perigo",
    deleteFunction: "Excluir Função",
    deleteWarning: "Uma vez excluída, esta função não pode ser recuperada.",
    deleteConfirm:
      'Tem certeza que deseja excluir "{{name}}"? Esta ação não pode ser desfeita.',
    retention: {
      days7: "7 dias",
      days15: "15 dias",
      days30: "30 dias",
      year1: "1 ano",
    },
  },

  // Executions
  executions: {
    title: "Histórico de Execuções",
    totalCount: "{{count}} execuções no total",
    emptyState:
      "Nenhuma execução ainda. Teste sua função para ver o histórico de execuções.",
    columns: {
      id: "ID da Execução",
      status: "Status",
      duration: "Duração",
      time: "Hora",
    },
  },

  // Test page
  test: {
    response: "Resposta",
    status: "Status",
    body: "Corpo",
    logs: "Logs",
    noResponse: "Nenhuma resposta ainda",
    executeHint: "Execute uma requisição para ver a resposta",
    viewExecution: "Ver Execução",
  },

  // Command palette
  commandPalette: {
    searchPlaceholder: "Buscar funções...",
    loading: "Carregando...",
    noResults: "Nenhum resultado encontrado",
    toNavigate: "para navegar",
    toSelect: "para selecionar",
    toClose: "para fechar",
    actions: {
      viewFunctions: "Ver todas as funções",
      createFunction: "Criar uma nova função",
      goToCode: "Ir para Código",
      viewVersions: "Ver histórico de versões",
      viewExecutions: "Ver logs de execução",
      configureFunction: "Configurar função",
      testFunction: "Testar função",
      switchLanguage: "Mudar idioma",
    },
    currentLanguage: "(atual)",
  },

  // Toast notifications
  toast: {
    closeNotification: "Fechar notificação",
    envVarsUpdated: "Variáveis de ambiente atualizadas",
    settingsSaved: "Configurações salvas com sucesso",
    functionDeleted: "Função excluída com sucesso",
    functionEnabled: "Função ativada com sucesso",
    functionDisabled: "Função desativada com sucesso",
    failedToSave: "Falha ao salvar configurações",
    failedToDelete: "Falha ao excluir função",
    failedToUpdate: "Falha ao atualizar status",
    executionFailed: "Execução falhou",
  },

  // Pagination
  pagination: {
    showing: "Mostrando",
    to: "até",
    of: "de",
    results: "resultados",
    perPage: "{{count}} por página",
    previous: "Anterior",
    next: "Próximo",
  },

  // Versions
  versions: {
    title: "Histórico de Versões",
    totalCount: "{{count}} versões no total",
    emptyState: "Nenhuma versão ainda.",
    current: "Atual",
    columns: {
      version: "Versão",
      createdAt: "Criado em",
      actions: "Ações",
    },
    compare: "Comparar",
    deploy: "Implantar",
  },

  // Diff viewer
  diff: {
    title: "Comparação de Versões",
    comparing: "Comparando",
    with: "com",
    addition: "adição",
    additions: "adições",
    deletion: "remoção",
    deletions: "remoções",
    codeDiff: "Diferença de código",
    lineAdded: "Linha adicionada",
    lineRemoved: "Linha removida",
    unchangedLine: "Linha inalterada",
  },

  // Request builder
  requestBuilder: {
    request: "Requisição",
    method: "Método",
    url: "URL",
    requestUrl: "URL da requisição",
    queryParams: "Parâmetros de Query",
    headers: "Headers (JSON)",
    requestBody: "Corpo da Requisição",
    execute: "Enviar Requisição",
    executing: "Enviando...",
  },

  // Badge
  badge: {
    enabled: "Ativa",
    disabled: "Desativa",
  },

  // Code editor
  code: {
    codeSaved: "Código salvo com sucesso",
    failedToSave: "Falha ao salvar código",
  },

  // Execution detail
  execution: {
    loadingExecution: "Carregando execução...",
    executionNotFound: "Execução não encontrada",
    executionError: "Erro de Execução",
    inputEvent: "Evento de Entrada (JSON)",
    aiRequests: "Requisições de IA",
    aiRequestsCount: "{{count}} chamadas de API",
    emailRequests: "Requisições de Email",
    emailsSent: "{{count}} emails enviados",
    executionLogs: "Logs de Execução",
    logEntries: "{{count}} entradas de log",
  },

  // Version diff
  versionDiff: {
    loadingDiff: "Carregando diff...",
    diffNotFound: "Diff não encontrado",
    codeChanges: "Alterações no Código",
  },

  // Environment variables
  envVars: {
    noVariables:
      'Sem variáveis de ambiente. Clique em "Adicionar Variável" para criar uma.',
    addVariable: "Adicionar Variável",
    keyPlaceholder: "CHAVE",
    valuePlaceholder: "Valor",
    restore: "Restaurar",
    remove: "Remover",
  },

  // Versions
  versionsPage: {
    activateConfirm: "Ativar versão {{version}}?",
    versionActivated: "Versão {{version}} ativada",
    failedToActivate: "Falha ao ativar versão",
    active: "ATIVA",
    activate: "Ativar",
    selectToCompare: "Selecione 2 versões para comparar",
    compareVersions: "Comparar v{{v1}} e v{{v2}}",
    versionsCount: "{{count}} versões",
  },

  // AI Request viewer
  aiRequestViewer: {
    noRequests: "Nenhuma requisição de IA registrada para esta execução.",
    provider: "Provedor",
    model: "Modelo",
    status: "Status",
    tokens: "Tokens",
    duration: "Duração",
    time: "Hora",
    in: "entrada",
    out: "saída",
    error: "Erro",
    endpoint: "Endpoint",
    request: "Requisição",
    response: "Resposta",
    truncated: "... (truncado)",
  },

  // API Reference
  apiReference: {
    llmDocumentation: "Documentação LLM",
  },

  // Card
  card: {
    maximize: "Maximizar",
    minimize: "Minimizar",
  },

  // Code examples
  codeExamples: {
    title: "Exemplos de Código",
    subtitle: "Chame esta função a partir da sua aplicação",
    copied: "Copiado!",
    copyToClipboard: "Copiar para área de transferência",
    selectLanguage: "a linguagem seleciona é",
  },

  // Email request viewer
  emailRequestViewer: {
    noRequests: "Nenhuma requisição de email registrada para esta execução.",
    to: "Para",
    subject: "Assunto",
    status: "Status",
    type: "Tipo",
    duration: "Duração",
    time: "Hora",
    error: "Erro",
    from: "De",
    emailId: "ID do Email",
    request: "Requisição",
    response: "Resposta",
    truncated: "... (truncado)",
  },

  // Form
  form: {
    showPassword: "Mostrar senha",
    hidePassword: "Ocultar senha",
    copied: "Copiado!",
    copyToClipboard: "Copiar para área de transferência",
    checkBox: "checkbox",
  },

  // Log viewer
  logViewer: {
    noLogs: "Nenhum log disponível",
  },

  // Table
  table: {
    noData: "Nenhum dado disponível",
  },

  // Lua API Reference
  luaApi: {
    types: {
      string: "string",
      number: "número",
      table: "tabela",
      function: "função",
      module: "módulo",
    },
    ai: {
      name: "IA",
      description: "Integrações com provedores de IA",
      groups: { chat: "Chat (IA)" },
      items: { chat: "Chat com OpenAI ou Anthropic" },
    },
    email: {
      name: "Email",
      description: "Envio de email via Resend",
      groups: { send: "Enviar (email)" },
      items: { send: "Enviar email via API Resend" },
    },
    handler: {
      name: "Handler",
      description: "Entradas da função handler",
      groups: {
        context: "Contexto (ctx)",
        event: "Evento (event)",
      },
      items: {
        executionId: "Identificador único da execução",
        functionId: "Identificador da função",
        functionName: "Nome da função",
        version: "Versão da função",
        requestId: "Identificador da requisição HTTP",
        startedAt: "Timestamp de início (Unix)",
        baseUrl: "URL base do servidor",
        method: "Método HTTP (GET, POST, etc.)",
        path: "Caminho da requisição",
        body: "Corpo da requisição como string",
        headers: "Tabela de cabeçalhos da requisição",
        query: "Tabela de parâmetros de query",
      },
    },
    io: {
      name: "IO",
      description: "Operações de entrada/saída",
      groups: {
        logging: "Logging (log)",
        kv: "Armazenamento Chave-Valor (kv)",
        env: "Ambiente (env)",
        http: "Cliente HTTP (http)",
      },
      items: {
        logInfo: "Registrar mensagem de info",
        logDebug: "Registrar mensagem de debug",
        logWarn: "Registrar mensagem de aviso",
        logError: "Registrar mensagem de erro",
        kvGet: "Obter valor do armazenamento",
        kvSet: "Definir par chave-valor",
        kvDelete: "Excluir chave do armazenamento",
        envGet: "Obter variável de ambiente",
        httpGet: "Requisição GET",
        httpPost: "Requisição POST",
        httpPut: "Requisição PUT",
        httpDelete: "Requisição DELETE",
      },
    },
    data: {
      name: "Dados",
      description: "Transformação de dados",
      groups: {
        json: "JSON (json)",
        base64: "Base64 (base64)",
        crypto: "Criptografia (crypto)",
      },
      items: {
        jsonEncode: "Codificar tabela para JSON",
        jsonDecode: "Decodificar JSON para tabela",
        base64Encode: "Codificar para base64",
        base64Decode: "Decodificar de base64",
        md5: "Hash MD5 (hex)",
        sha256: "Hash SHA256 (hex)",
        hmacSha256: "HMAC-SHA256 (hex)",
        uuid: "Gerar UUID v4",
      },
    },
    utils: {
      name: "Utils",
      description: "Funções utilitárias",
      groups: {
        time: "Tempo (time)",
        strings: "Strings (strings)",
        random: "Aleatório (random)",
      },
      items: {
        timeNow: "Timestamp Unix atual",
        timeFormat: "Formatar timestamp",
        timeParse: "Converter texto em data",
        timeSleep: "Pausar milissegundos",
        trim: "Remover espaços",
        split: "Dividir por separador",
        join: "Juntar com separador",
        contains: "Verificar se contém texto",
        replace: "Substituir na string",
        randomInt: "Inteiro aleatório",
        randomFloat: "Float aleatório 0.0-1.0",
        randomString: "String alfanumérica aleatória",
        randomId: "ID único ordenável",
      },
    },
  },
};
