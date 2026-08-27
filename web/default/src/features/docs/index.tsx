import {
  BookOpen,
  CheckCircle2,
  Code2,
  CreditCard,
  Download,
  KeyRound,
  MessageSquare,
  Monitor,
  Smartphone,
  type LucideIcon,
} from 'lucide-react'
import { Main } from '@/components/layout'

type TextBlock = {
  kind: 'text'
  text: string
}

type ImageBlock = {
  kind: 'images'
  images: string[]
}

type TutorialBlock = TextBlock | ImageBlock

type TutorialSection = {
  id: string
  title: string
  icon: LucideIcon
  blocks: TutorialBlock[]
}

const imageSrc = (name: string) => `/docs/tutorial/${name}`

const tutorialSections: TutorialSection[] = [
  {
    id: 'purchase',
    title: '1 购买额度',
    icon: CreditCard,
    blocks: [
      { kind: 'images', images: ['image1.png'] },
      {
        kind: 'text',
        text: '步骤 1.1：在概览页面右下角点击“购买兑换码”。系统将跳转至闲鱼商店，下单后会自动发放兑换码。目前提供 10 元面额兑换码。',
      },
      {
        kind: 'text',
        text: '步骤 1.2：红框中显示的内容即为兑换码。请妥善保存兑换码，以便后续完成额度兑换或账户配置。',
      },
      { kind: 'images', images: ['image3.png'] },
    ],
  },
  {
    id: 'downloads',
    title: '2 软件下载',
    icon: Download,
    blocks: [
      {
        kind: 'text',
        text: '步骤 2.1：普通对话用户请下载并安装 Cherry Studio，用于日常聊天和模型调用。',
      },
      {
        kind: 'text',
        text: '步骤 2.2：代码编程用户请下载 CC Switch 作为转换接口。仅进行普通对话的用户可跳过 CC Switch 相关配置。',
      },
      {
        kind: 'text',
        text: '步骤 2.3：在页面中点击“Coding”，进入编程工具下载区域。',
      },
      { kind: 'images', images: ['image4.png'] },
      {
        kind: 'text',
        text: '步骤 2.4：根据当前电脑系统下载对应版本的 CC Switch。',
      },
      {
        kind: 'text',
        text: '步骤 2.5：编程工具建议使用 VS Code，并在插件商店中安装 OpenCode、Codex 或 Claude Code 等扩展。其中 Codex 适配性较好，推荐优先使用。',
      },
      {
        kind: 'text',
        text: '说明：当前主要支持 GPT 模型，因此建议选择兼容性更好的 Codex。',
      },
    ],
  },
  {
    id: 'api-key',
    title: '3 API 密钥创建与使用',
    icon: KeyRound,
    blocks: [
      {
        kind: 'text',
        text: '步骤 3.1：在“概览”页面右侧找到“API 密钥”区域，并点击创建密钥。',
      },
      { kind: 'images', images: ['image5.png', 'image6.png', 'image7.png'] },
      {
        kind: 'text',
        text: '步骤 3.2：按图示点击右侧三个点，展开更多操作选项。',
      },
      { kind: 'images', images: ['image8.png'] },
    ],
  },
  {
    id: 'cherry-studio',
    title: '4 Cherry Studio 导入与使用',
    icon: MessageSquare,
    blocks: [
      {
        kind: 'text',
        text: '步骤 4.1：在控制台中点击“聊天”，再点击“Cherry Studio”。操作前请确认电脑已安装 Cherry Studio。',
      },
      { kind: 'images', images: ['image9.png'] },
      { kind: 'text', text: '步骤 4.2：打开 Cherry Studio 客户端。' },
      { kind: 'images', images: ['image10.png', 'image11.png'] },
      {
        kind: 'text',
        text: '步骤 4.3：点击“添加”，准备新增服务或模型配置。',
      },
      { kind: 'images', images: ['image12.png'] },
      {
        kind: 'text',
        text: '步骤 4.4：点击“获取模型列表”，等待模型信息加载完成。',
      },
      { kind: 'images', images: ['image13.png'] },
      {
        kind: 'text',
        text: '步骤 4.5：在需要使用的模型右侧点击“+”，将该模型添加到列表。',
      },
      { kind: 'images', images: ['image14.png'] },
      {
        kind: 'text',
        text: '步骤 4.6：添加完成后，可在模型列表中查看已添加模型，效果如下图所示。',
      },
      { kind: 'images', images: ['image15.png'] },
      {
        kind: 'text',
        text: '步骤 4.7：在对应助手右侧点击模型切换按钮，选择刚刚添加的对话模型。',
      },
      { kind: 'images', images: ['image16.png'] },
      { kind: 'text', text: '4.8 对话输入区域说明' },
      {
        kind: 'text',
        text: '4.8.1 “思维链长度”会影响模型的思考深度、响应耗时以及 token 消耗。',
      },
      {
        kind: 'text',
        text: '4.8.2 通常建议保持默认设置，由模型根据问题复杂度自动判断。',
      },
      {
        kind: 'text',
        text: '4.8.3 思维链长度越长，模型思考时间越长，token 消耗越多，回答时考虑的信息也会更充分。',
      },
      { kind: 'images', images: ['image17.png', 'image18.png'] },
      {
        kind: 'text',
        text: '4.8.4 网络搜索功能保持默认关闭即可。当前模型默认具备联网能力，一般无需手动开启。',
      },
      { kind: 'images', images: ['image19.png'] },
    ],
  },
  {
    id: 'rikkahub',
    title: '5 RikkaHub 手机端导入与使用',
    icon: Smartphone,
    blocks: [
      {
        kind: 'text',
        text: '步骤 5.1：在 RikkaHub 中点击箭头处的设置，进入配置页面。',
      },
      { kind: 'images', images: ['rikkahub/image7.png'] },
      { kind: 'text', text: '步骤 5.2：点击“提供商”。' },
      { kind: 'images', images: ['rikkahub/image6.png'] },
      { kind: 'text', text: '步骤 5.3：找到并打开 newapi 条目。' },
      { kind: 'images', images: ['rikkahub/image5.png'] },
      {
        kind: 'text',
        text: '步骤 5.4：填写 API 密钥和 Base URL，勾选 Response API 并点击保存。之后点击右下角“模型”。',
      },
      { kind: 'images', images: ['rikkahub/image4.png'] },
      {
        kind: 'text',
        text: '步骤 5.5：点击箭头处图标，打开模型选择界面。',
      },
      { kind: 'images', images: ['rikkahub/image3.png'] },
      {
        kind: 'text',
        text: '步骤 5.6：点击右侧“+”添加模型，再点击左下角“配置”返回配置主界面并点击保存。',
      },
      { kind: 'images', images: ['rikkahub/image2.png'] },
      {
        kind: 'text',
        text: '步骤 5.7：在“新聊天”中点击箭头所示图标选择对话使用的模型，即可开始对话。',
      },
      { kind: 'images', images: ['rikkahub/image1.png'] },
    ],
  },
  {
    id: 'cc-switch',
    title: '6 CC Switch 导入配置',
    icon: Monitor,
    blocks: [
      { kind: 'images', images: ['image20.png'] },
      { kind: 'text', text: '步骤 6.1：选择需要使用的主模型。' },
      { kind: 'images', images: ['image21.png'] },
      { kind: 'text', text: '步骤 6.2：确认模型选择无误后，继续下一步配置。' },
      { kind: 'text', text: '步骤 6.3：打开 CC Switch。' },
      { kind: 'images', images: ['image22.png'] },
      { kind: 'images', images: ['image23.png', 'image24.png'] },
      {
        kind: 'text',
        text: '步骤 6.4：点击“导入”。页面提示添加成功后，即表示 CC Switch 配置已完成。',
      },
      { kind: 'images', images: ['image25.png'] },
      { kind: 'images', images: ['image26.png'] },
      { kind: 'images', images: ['image27.png'] },
    ],
  },
  {
    id: 'vscode-codex',
    title: '7 VS Code / Codex 使用',
    icon: Code2,
    blocks: [
      { kind: 'images', images: ['image28.png'] },
      {
        kind: 'text',
        text: '步骤 7.1：打开 VS Code 后，点击图中箭头所示的“切换辅助侧栏”，唤出 Codex 面板。若未显示 Codex，请先确认已安装 Codex 扩展。',
      },
      { kind: 'images', images: ['image29.png'] },
      { kind: 'images', images: ['image30.png'] },
      {
        kind: 'text',
        text: '步骤 7.2：底部显示“my_codex”时，表示当前配置已切换完成，可以开始使用。',
      },
    ],
  },
]

export function Docs() {
  return (
    <Main>
      <div className='min-h-0 flex-1 overflow-auto px-3 py-3 sm:px-4 sm:py-6'>
        <article className='mx-auto flex w-full max-w-6xl flex-col gap-5 sm:gap-6'>
          <header className='border-border/70 rounded-lg border bg-card px-5 py-5 sm:px-6'>
            <div className='space-y-4'>
              <div className='text-muted-foreground flex items-center gap-2 text-sm font-medium'>
                <BookOpen className='size-4' />
                <span>使用教程</span>
              </div>
              <div className='space-y-2'>
                <h1 className='text-2xl font-semibold tracking-tight sm:text-3xl'>
                  API 创建与使用教程
                </h1>
                <p className='text-muted-foreground max-w-4xl text-sm leading-6 sm:text-base'>
                  本教程面向需要通过 API 进行普通对话或代码编程的用户，按照“购买额度、下载软件、创建密钥、导入电脑端或手机端客户端、在编程工具中使用”的顺序进行说明。请按章节编号依次操作。
                </p>
              </div>
            </div>
          </header>

          <nav className='rounded-lg border bg-card p-4 sm:p-5'>
            <div className='mb-3 flex items-center gap-2'>
              <CheckCircle2 className='text-primary size-5' />
              <h2 className='text-lg font-semibold'>目录</h2>
            </div>
            <div className='grid gap-2 sm:grid-cols-2 xl:grid-cols-3'>
              {tutorialSections.map((section) => {
                const Icon = section.icon
                return (
                  <a
                    key={section.id}
                    href={`#${section.id}`}
                    className='hover:border-primary/60 hover:bg-muted/50 flex items-center gap-3 rounded-md border px-3 py-3 transition-colors'
                  >
                    <Icon className='text-primary size-4 shrink-0' />
                    <span className='text-sm font-medium'>{section.title}</span>
                  </a>
                )
              })}
            </div>
          </nav>

          <div className='space-y-5 sm:space-y-6'>
            {tutorialSections.map((section) => (
              <TutorialSectionCard key={section.id} section={section} />
            ))}
          </div>
        </article>
      </div>
    </Main>
  )
}

function TutorialSectionCard(props: { section: TutorialSection }) {
  const Icon = props.section.icon

  return (
    <section
      id={props.section.id}
      className='scroll-mt-16 rounded-lg border bg-card p-4 sm:p-6'
    >
      <div className='mb-5 flex items-center gap-2'>
        <Icon className='text-primary size-5' />
        <h2 className='text-xl font-semibold tracking-tight'>
          {props.section.title}
        </h2>
      </div>

      <div className='space-y-4'>
        {props.section.blocks.map((block, index) => {
          if (block.kind === 'text') {
            return (
              <p
                key={`${props.section.id}-text-${index}`}
                className='text-muted-foreground text-sm leading-7 sm:text-base'
              >
                {block.text}
              </p>
            )
          }

          return (
            <TutorialImages
              key={`${props.section.id}-images-${index}`}
              images={block.images}
              sectionTitle={props.section.title}
            />
          )
        })}
      </div>
    </section>
  )
}

function TutorialImages(props: { images: string[]; sectionTitle: string }) {
  const gridClass =
    props.images.length === 1
      ? 'grid-cols-1'
      : 'grid-cols-1 md:grid-cols-2 xl:grid-cols-3'

  return (
    <div className={`grid gap-3 ${gridClass}`}>
      {props.images.map((image) => (
        <figure
          key={image}
          className='bg-muted/20 overflow-hidden rounded-lg border'
        >
          <img
            src={imageSrc(image)}
            alt={`${props.sectionTitle} 截图 ${image.replace('image', '').replace('.png', '')}`}
            loading='lazy'
            className='max-h-[620px] w-full object-contain'
          />
        </figure>
      ))}
    </div>
  )
}
