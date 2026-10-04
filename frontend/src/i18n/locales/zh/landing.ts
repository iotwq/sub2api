export default {
  batchImageGuide: {
    title: '图片批量生成',
    description: '一次提交多条提示词，任务完成后可统一下载图片结果'
  },
  // Home Page
  home: {
    viewOnGithub: '在 GitHub 上查看',
    viewDocs: '查看文档',
    docs: '文档',
    switchToLight: '切换到浅色模式',
    switchToDark: '切换到深色模式',
    dashboard: '控制台',
    login: '登录',
    getStarted: '立即开始',
    goToDashboard: '进入控制台',
    // 新增：面向用户的价值主张
    heroSubtitle: 'AI技术 学习交流',
    heroDescription: '一起探索大模型、Agent、Prompt 与 AI 应用实践，连接更多前沿想法',
    agent: {
      eyebrow: 'AI AGENT RESEARCH NETWORK',
      headline: '研究真正能执行任务的 AI Agent',
      explore: '查看研究方向',
      sceneLabel: '实时 AI Agent 编排网络',
      core: 'ORCHESTRATOR',
      coreStatus: '自主循环运行中',
      nodes: {
        model: '推理模型',
        modelDetail: '理解 · 规划 · 决策',
        tools: '工具系统',
        toolsDetail: '搜索 · 代码 · API',
        memory: '长期记忆',
        memoryDetail: '检索 · 上下文 · 经验',
        evaluator: '结果评估',
        evaluatorDetail: '验证 · 反思 · 修正'
      },
      traceEyebrow: 'LIVE AGENT TRACE',
      traceTitle: '一次任务，四个可验证阶段',
      steps: {
        plan: '任务规划',
        reason: '模型推理',
        act: '工具执行',
        verify: '结果校验'
      },
      metrics: {
        tools: '可用工具',
        context: '上下文',
        active: '运行中'
      },
      researchEyebrow: 'RESEARCH TRACKS / 2026',
      researchTitle: '从模型能力到可运行的智能体系统',
      researchDescription: '关注真实能力边界、工程方法与可复现的 AI 前沿实践。',
      tracks: {
        systems: 'Agent 系统',
        systemsDetail: '规划、记忆、协作与自主任务循环',
        multimodal: '多模态智能',
        multimodalDetail: '文本、图像、语音与视频的统一理解',
        tooling: '工具与协议',
        toolingDetail: 'MCP、函数调用、代码执行与工作流编排',
        evaluation: '评估与对齐',
        evaluationDetail: '用可验证指标复盘可靠性与真实效果'
      }
    },
    heroPanel: {
      eyebrow: 'AI Learning Lab',
      title: '知识流正在同步',
      live: '实时共创',
      stream1Label: '模型理解',
      stream1Value: '深度拆解',
      stream2Label: 'Prompt 实验',
      stream2Value: '灵感迭代',
      stream3Label: 'Agent 实践',
      stream3Value: '案例复盘',
      footer: '学习节点已连接'
    },
    tags: {
      aiFrontier: '前沿 AI 技术',
      deepExchange: '深度交流',
      promptLab: 'Prompt 实验场'
    },
    // 用户痛点区块
    painPoints: {
      title: 'AI 学习中，你是否也遇到这些问题？',
      items: {
        expensive: {
          title: '信息太碎片',
          desc: '模型、Agent、Prompt 的新概念不断出现，很难快速抓住重点'
        },
        complex: {
          title: '实践路径不清',
          desc: '看过很多教程，却不知道怎样把方法落到自己的项目里'
        },
        unstable: {
          title: '经验难复用',
          desc: '一次调通的 Prompt 或工作流，换个场景又需要重新摸索'
        },
        noControl: {
          title: '交流噪音高',
          desc: '真正有启发的案例、复盘和技术讨论，常常被信息流淹没'
        }
      }
    },
    // 解决方案区块
    solutions: {
      title: '一起把 AI 学明白、用起来',
      subtitle: '围绕前沿技术、真实案例和可复用方法持续交流'
    },
    features: {
      unifiedGateway: '知识入口',
      unifiedGatewayDesc: '围绕大模型、Agent、Prompt 和 AI 应用实践，整理清晰的学习线索。',
      multiAccount: '实践共创',
      multiAccountDesc: '把真实问题、方案尝试和工具链经验放到同一个交流场里持续迭代。',
      balanceQuota: '案例沉淀',
      balanceQuotaDesc: '把有效的提示词、流程和复盘留下来，让下一次探索更快开始。',
      frontierLearning: '前沿技术共学',
      frontierLearningDesc: '聚焦大模型、生成式 AI、多模态和 Agent 最新进展，把复杂概念拆成可理解的知识路径。',
      communityExchange: '高质量交流',
      communityExchangeDesc: '围绕真实问题讨论思路、工具链和实践经验，让灵感在持续碰撞中变得更清晰。',
      promptLab: 'Prompt 实验室',
      promptLabDesc: '沉淀提示词、工作流和案例复盘，用可复制的方法提升 AI 应用和创作效率。'
    },
    // 优势对比
    comparison: {
      title: '为什么适合 AI 技术交流？',
      headers: {
        feature: '对比项',
        official: '碎片自学',
        us: 'ChinaAPI 共学'
      },
      items: {
        pricing: {
          feature: '学习路径',
          official: '信息分散，难以串联',
          us: '主题聚合，路径更清晰'
        },
        models: {
          feature: '技术视野',
          official: '跟着单点资料学习',
          us: '覆盖模型、Agent 与应用案例'
        },
        management: {
          feature: '实践沉淀',
          official: '看完容易遗忘',
          us: '复盘方法和可复用经验'
        },
        stability: {
          feature: '问题讨论',
          official: '独自试错成本高',
          us: '围绕真实场景共同拆解'
        },
        control: {
          feature: '灵感来源',
          official: '灵感零散不可追踪',
          us: '案例、Prompt、工作流持续更新'
        }
      }
    },
    providers: {
      title: 'AI 技术学习主题',
      description: '从模型认知到应用实践',
      supported: '热门',
      soon: '持续更新',
      claude: '模型思维',
      gemini: 'Prompt 工程',
      antigravity: 'Agent 工作流',
      more: '更多主题'
    },
    topics: {
      title: 'AI 技术学习主题',
      description: '从模型认知到应用实践，让交流更聚焦、更有启发',
      hot: '热门',
      new: '新方向',
      soon: '持续更新',
      modelThinking: '模型思维',
      promptEngineering: 'Prompt 工程',
      agentWorkflow: 'Agent 工作流',
      caseStudies: '实战案例',
      more: '更多主题'
    },
    // CTA 区块
    cta: {
      title: '准备好开始探索了吗？',
      description: '进入 ChinaAPI，一起学习、交流和实践最新 AI 技术',
      button: '立即开始'
    },
    footer: {
      allRightsReserved: '保留所有权利。'
    }
  },

  // Key Usage Query Page
  keyUsage: {
    title: 'API Key 用量查询',
    subtitle: '输入您的 API Key 以查看实时消费金额与使用状态',
    placeholder: 'sk-ant-mirror-xxxxxxxxxxxx',
    query: '查询',
    querying: '查询中...',
    privacyNote: '您的 Key 仅在浏览器本地处理，不会被存储',
    dateRange: '统计范围:',
    dateRangeToday: '今日',
    dateRange7d: '7 天',
    dateRange30d: '30 天',
    dateRange90d: '90 天',
    dateRangeCustom: '自定义',
    apply: '应用',
    used: '已使用',
    detailInfo: '详细信息',
    tokenStats: 'Token 统计',
    dailyDetail: '按日明细',
    modelStats: '模型用量统计',
    // Table headers
    date: '日期',
    model: '模型',
    requests: '请求数',
    inputTokens: '输入 Tokens',
    outputTokens: '输出 Tokens',
    cacheCreationTokens: '缓存创建',
    cacheReadTokens: '缓存读取',
    cacheWriteTokens: '缓存写入',
    totalTokens: '总 Tokens',
    cost: '费用',
    // Status
    quotaMode: 'Key 限额模式',
    walletBalance: '钱包余额',
    // Ring card titles
    totalQuota: '总额度',
    limit5h: '5 小时限额',
    limitDaily: '日限额',
    limit7d: '7 天限额',
    limitWeekly: '周限额',
    limitMonthly: '月限额',
    // Detail rows
    remainingQuota: '剩余额度',
    expiresAt: '过期时间',
    todayExpires: '(今日到期)',
    daysLeft: '({days} 天)',
    usedQuota: '已用额度',
    resetNow: '即将重置',
    subscriptionType: '订阅类型',
    billingType: '计费方式',
    subscriptionExpires: '订阅到期',
    // Usage stat cells
    todayRequests: '今日请求',
    todayInputTokens: '今日输入',
    todayOutputTokens: '今日输出',
    todayTokens: '今日 Tokens',
    todayCacheCreation: '今日缓存创建',
    todayCacheRead: '今日缓存读取',
    todayCost: '今日费用',
    rpmTpm: 'RPM / TPM',
    totalRequests: '累计请求',
    totalInputTokens: '累计输入',
    totalOutputTokens: '累计输出',
    totalTokensLabel: '累计 Tokens',
    totalCacheCreation: '累计缓存创建',
    totalCacheRead: '累计缓存读取',
    totalCost: '累计费用',
    avgDuration: '平均耗时',
    // Messages
    enterApiKey: '请输入 API Key',
    querySuccess: '查询成功',
    queryFailed: '查询失败',
    queryFailedRetry: '查询失败，请稍后重试',
    noDailyUsage: '暂无按日用量数据',
  },

  // Setup Wizard
  setup: {
    title: 'Sub2API 安装向导',
    description: '配置您的 Sub2API 实例',
    database: {
      title: '数据库配置',
      description: '连接到您的 PostgreSQL 数据库',
      host: '主机',
      port: '端口',
      username: '用户名',
      password: '密码',
      databaseName: '数据库名称',
      sslMode: 'SSL 模式',
      passwordPlaceholder: '密码',
      ssl: {
        disable: '禁用',
        require: '要求',
        verifyCa: '验证 CA',
        verifyFull: '完全验证'
      }
    },
    redis: {
      title: 'Redis 配置',
      description: '连接到您的 Redis 服务器',
      host: '主机',
      port: '端口',
      username: '用户名（可选）',
      password: '密码（可选）',
      database: '数据库',
      usernamePlaceholder: '默认用户留空',
      passwordPlaceholder: '密码',
      enableTls: '启用 TLS',
      enableTlsHint: '连接 Redis 时使用 TLS（公共 CA 证书）'
    },
    admin: {
      title: '管理员账户',
      description: '创建您的管理员账户',
      email: '邮箱',
      password: '密码',
      confirmPassword: '确认密码',
      passwordPlaceholder: '至少 8 个字符',
      confirmPasswordPlaceholder: '确认密码',
      passwordMismatch: '密码不匹配'
    },
    ready: {
      title: '准备安装',
      description: '检查您的配置并完成安装',
      database: '数据库',
      redis: 'Redis',
      adminEmail: '管理员邮箱'
    },
    status: {
      testing: '测试中...',
      success: '连接成功',
      testConnection: '测试连接',
      installing: '安装中...',
      completeInstallation: '完成安装',
      completed: '安装完成！',
      redirecting: '正在跳转到登录页面...',
      restarting: '服务正在重启，请稍候...',
      timeout: '服务重启时间超出预期，请手动刷新页面。'
    }
  },

  // Common
}
