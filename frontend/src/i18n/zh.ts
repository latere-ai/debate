export const zh: Record<string, string> = {
  'nav.theme': '切换主题',
  'nav.install': '安装',

  'hero.eyebrow': '对抗式验证关卡',
  'hero.title': '让 AI 的产出<br>为自己辩护 —— <span class="agon-brand">Agon</span>',
  'hero.sub':
    '由一位独立评审对 AI 的产出进行交叉质询。生产方必须为每一条具体质疑辩护或让步。只有最终存留的争议才会交到人手中。',
  'hero.cta.install': '安装',
  'hero.cta.repo': '阅读协议',

  'what.label': '这是什么',
  'what.title': '不是裁判，而是一套<span class="agon-brand">协议</span>。',
  'what.p1':
    '<strong>Agon</strong> 位于 AI 智能体与其产出消费方（人或另一个智能体）之间。它用一位诚实、有能力的评审审视产出，迫使生产方为每一条具体质疑辩护或修正，并只呈现最终仍存争议的部分。',
  'what.p2':
    '被审对象不限于代码：一段补丁、一篇研究论述、一份结果分析、一项计划、一个高风险决策——同一套协议都适用。Agon 的内部代号为 <code>debate</code>；二进制、代码仓库与 Stop-hook 契约均以该名发布。',

  'proto.label': '协议',
  'proto.title': '四轮对话，外部裁定的争点',
  'proto.lead':
    '一个提议方与一个评审方；裁判只检视争论最终落到的那一个叶子节点，绝不通读完整记录。',
  'proto.s1.t': '提议',
  'proto.s1.d': '提议方完成任务——给出一个论断、一段补丁或一个论证。',
  'proto.s2.t': '攻击',
  'proto.s2.d': '评审方给出具体攻击：某个输入 X 产生输出 Y，违反了 Z。不接受空泛说辞。',
  'proto.s3.t': '辩护',
  'proto.s3.d': '提议方逐条回应：让步，或以具体的反论据进行反驳。',
  'proto.s4.t': '争点',
  'proto.s4.d': '评审方挑出一条未解决的攻击作为决定性叶子。若全部让步，提议方默认胜出。',
  'proto.judge':
    '裁判只评估被押注的那个叶子。可靠性不再要求诚实多数——只需一位诚实参与者加一位经过校准的裁判。这降低了信任预算，但并未消除信任。',

  'why.label': '为何不同',
  'why.title': '四个替代方案无法复制的性质',
  'why.c1.t': '一位诚实参与者即足够',
  'why.c1.d':
    '拜占庭式的提议方必须在每一轮交叉质询中维持一致的谎言；而诚实的评审只需找出一处不一致。失效模式因此变为按维度局部失效，而非整体失效。',
  'why.c2.t': '构造上厂商中立',
  'why.c2.d':
    '默认配对为跨家族——Claude 提议，Codex 评审。两侧同模型即“模型自我辩论”，被拒绝。没有模型厂商会去做这一中立层。',
  'why.c3.t': '通道纯净',
  'why.c3.d':
    '评审输出以逐字用户消息抵达提议方，而非技能或模板。提议方就像面对人类贴来的评审那样辩护。任何包装都会扭曲辩护行为。',
  'why.c4.t': '设计上可审计',
  'why.c4.d':
    '稳定的攻击 id、仅追加的账本、由纯规则给出争议度排序的头条——呈现层不做 LLM 裁决。安全团队可像读庭审记录一样读一次会话。',

  'signal.label': '已解决 vs. 仍存争议',
  'signal.title': '一个机器可读的信号，而不只是给人看的头条',
  'signal.p1':
    '在智能体回路内，争议度评分就是一道决策关卡：提议方已化解的攻击是调用方可据以继续的<strong>通过</strong>信号；高于阈值的争议尾部才升级给人。',
  'signal.p2':
    '这正是高风险决策的用法——智能体解决 ⇒ 继续；未解决 ⇒ 人工复核。这并非新增能力；当消费方是另一个智能体而非人时，既有账本与争议度评分本就是这个信号。',

  'found.label': '学术基础',
  'found.title': '扎根于 debate 文献——并诚实交代其边界',
  'found.p1':
    'Agon 将 Irving、Christiano 与 Amodei（2018）的对抗式辩论架构产品化。其复杂度理论直觉——最优博弈下 debate ≈ PSPACE，严格高于 NP——只是<em>启发性的</em>，并非针对 LLM 的论断：LLM 并非最优博弈者。',
  'found.p2':
    '更贴近的理论是 Brown-Cohen、Irving 与 Piliouras（2023），它将结果推广到随机系统、并将诚实辩手的模拟预算从指数降为多项式——二者都是该结果能适用于 LLM 的前提。其 2025 年的 Prover-Estimator 协议针对朴素 debate 的“混淆论证”攻击给出对策。',
  'found.research.cite': '研究主页。',
  'found.research.note':
    '本产品所产品化的实验套件（spec 07，对抗式辩论；spec 08–13 沿算力、深度、随机性、叶子格式、混淆、查询复杂度伸缩等六个维度扩展）。',
  'found.honest':
    '诚实表述：形式化可靠性结果是在既定假设下关于协议本身的，并非对任何具体模型的保证。应用到真实 LLM 是经验驱动、仍处假设验证阶段——关键门槛指标是各维度的“评审命中缺陷率”；若评审并不真正发起攻击，debate 便退化为只有提议方。Agon 不证明你的代码正确，也不声称免除信任的必要。',

  'install.label': '安装',
  'install.title': '一个二进制，一个可选的 Stop hook',
  'install.lead':
    '厂商中立、本地优先。自带 Claude / Codex；Agon 运行协议，并把可审计的会话写入磁盘。',
  'install.c1': '一行命令（检测 OS/架构、校验 checksum、安装 Stop hook）',
  'install.c2': '从源码（Go 1.26+）',
  'install.repo': '在 GitHub 查看',

  'footer.docs': '文档',
  'footer.research': '研究',
  'footer.tagline': '为 AI 产出而生的对抗式验证',

  'nf.text': '该页面不存在。',
  'nf.home': '返回首页',
};
