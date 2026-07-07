<template>
  <AppLayout>
    <div class="image-workbench">
      <!-- Floating Top Bar -->
      <div class="floating-topbar">
        <div class="topbar-left">
          <button class="topbar-btn" @click="isHistoryOpen = true">
            <Icon name="clock" size="sm" />
            <span class="topbar-btn-label">历史对话</span>
            <span class="topbar-badge">{{ conversations.length }}</span>
          </button>
          <button class="topbar-btn topbar-btn-primary" @click="newConversation">
            <Icon name="plus" size="sm" />
            <span class="topbar-btn-label">新建</span>
          </button>
          <button class="topbar-btn" @click="confirmClearAll = true" :disabled="conversations.length === 0" title="清空历史">
            <Icon name="trash" size="sm" />
          </button>
        </div>
        <div class="topbar-right">
          <div class="workbench-mode-switch" aria-label="GPTImage 模式">
            <button class="workbench-mode-tab workbench-mode-tab-active" type="button" @click="focusComposer">
              画廊
            </button>
            <button class="workbench-mode-tab" type="button" disabled title="Agent 模式即将接入">
              Agent
            </button>
          </div>
          <div class="workbench-control-strip" aria-label="GPTImage 控制区">
            <span class="workbench-status-chip" title="当前运行模式">
              <span class="workbench-status-dot"></span>
              {{ quotaDisplay }}
            </span>
            <button class="workbench-console-btn" type="button" title="打开控制台" @click="isHistoryOpen = true">
              <Icon name="terminal" size="sm" />
              <span class="workbench-console-label">控制台</span>
              <span class="workbench-count">{{ activeTaskCount }}</span>
            </button>
            <button class="workbench-icon-btn" type="button" title="操作指南" aria-label="操作指南" @click="showWorkbenchHelp = true">
              <Icon name="questionCircle" size="sm" />
            </button>
            <button class="workbench-icon-btn" type="button" title="生成设置" aria-label="生成设置" @click="showWorkbenchSettings = true">
              <Icon name="cog" size="sm" />
            </button>
          </div>
          <router-link class="workbench-works-link" to="/image-management" title="我的作品">
            <Icon name="grid" size="sm" />
            <span class="workbench-works-label">我的作品</span>
          </router-link>
        </div>
      </div>

      <!-- Results Area -->
      <div ref="resultsViewport" class="results-viewport" :class="{ 'results-viewport-empty': !activeConversation }">
        <template v-if="!activeConversation">
          <div class="empty-hero">
            <div class="empty-hero-eyebrow"><span></span>Generative · Atelier<span></span></div>
            <h1 class="empty-hero-title">Turn ideas into images</h1>
            <p class="empty-hero-sub">在同一窗口里保留本地历史与任务状态，并从已有结果图继续发起新的无状态编辑。</p>
            <div class="empty-hero-index"><span>01</span><span></span><span class="uppercase">Sketch → Render</span><span></span><span>02</span></div>
          </div>
        </template>
        <template v-else>
          <div class="conversation-turns">
            <div v-for="(turn, ti) in activeConversation.turns" :key="turn.id" class="turn">
              <!-- User prompt (right aligned bubble) -->
              <div v-if="!turn.promptDeleted" class="turn-prompt-row">
                <div class="turn-prompt-bubble">
                  <div class="turn-prompt-meta">
                    <span>第 {{ ti + 1 }} 轮</span>
                    <span>{{ turn.mode === 'edit' ? '编辑图' : '文生图' }}</span>
                    <span>{{ formatTime(turn.createdAt) }}</span>
                  </div>
                  <div class="turn-prompt-text">{{ turn.prompt }}</div>
                  <div class="turn-prompt-actions">
                    <button class="turn-chip" @click="reuseConfig(turn)">复用配置</button>
                    <button class="turn-icon-btn" @click="deleteTurnPrompt(turn.id)" title="删除提示词"><Icon name="trash" size="xs" /></button>
                  </div>
                </div>
              </div>

              <!-- AI results (left aligned) -->
              <div v-if="!turn.resultsDeleted" class="turn-results-row">
                <div class="turn-results-block">
                  <!-- Reference images -->
                  <div v-if="turn.referenceImages?.length" class="turn-ref-row">
                    <span class="turn-ref-label">本轮参考图</span>
                    <div class="turn-ref-images">
                      <div v-for="(ref, ri) in turn.referenceImages" :key="ri" class="turn-ref-item">
                        <img :src="ref.dataUrl" class="turn-ref-img" @click="openLightbox(turn, ri)" />
                        <button class="turn-ref-edit-btn" @click="continueEdit(turn, ref)"><Icon name="sparkles" size="xs" />加入编辑</button>
                      </div>
                    </div>
                  </div>

                  <!-- Image grid -->
                  <div v-if="turn.images.length" class="turn-image-grid">
                    <div v-for="(img, ii) in turn.images" :key="img.id" class="turn-image-card" :class="{ 'turn-image-loading': img.status === 'loading', 'turn-image-error': img.status === 'error' }">
                      <template v-if="img.status === 'success' && img.src">
                        <img :src="img.src" class="turn-image-preview" @click="openLightbox(turn, ii)" />
                        <div class="turn-image-footer">
                          <span class="turn-image-label">结果 {{ ii + 1 }}</span>
                          <span v-if="img.sizeLabel" class="turn-image-meta">{{ img.sizeLabel }}</span>
                          <div class="turn-image-actions">
                            <button class="turn-icon-btn" @click="continueEdit(turn, img)" title="加入编辑"><Icon name="sparkles" size="xs" /></button>
                            <button class="turn-icon-btn" @click="downloadImage(img, ii)" title="下载"><Icon name="download" size="xs" /></button>
                          </div>
                        </div>
                      </template>
                      <template v-else-if="img.status === 'loading'">
                        <div class="turn-image-loader">
                          <Icon name="refresh" size="lg" class="animate-spin text-stone-400" />
                          <span>{{ turn.status === 'queued' ? '排队中' : '处理中' }}</span>
                        </div>
                      </template>
                      <template v-else-if="img.status === 'error'">
                        <div class="turn-image-error-card">
                          <Icon name="exclamationCircle" size="md" />
                          <p class="turn-error-msg">{{ img.error || '生成失败' }}</p>
                          <div class="turn-error-actions">
                            <button class="turn-chip-primary" @click="retryImage(turn.id, img)">重试</button>
                            <button class="turn-chip" @click="retryImage(turn.id, img)">回复</button>
                            <button class="turn-icon-btn turn-error-delete-btn" @click="deleteTurnResults(turn.id)" title="删除结果" aria-label="删除结果">
                              <Icon name="trash" size="xs" />
                            </button>
                          </div>
                        </div>
                      </template>
                    </div>
                  </div>

                  <!-- Turn footer -->
                  <div v-if="turn.status !== 'queued' && turn.status !== 'generating' && !turn.error" class="turn-footer">
                    <button class="turn-chip" @click="regenerateTurn(turn)">全部重新生成</button>
                    <button class="turn-icon-btn" @click="deleteTurnResults(turn.id)" title="删除结果"><Icon name="trash" size="xs" /></button>
                  </div>
                  <div v-if="turn.error && turn.status === 'error'" class="turn-error-banner">
                    <Icon name="exclamationCircle" size="sm" /><span>{{ turn.error }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- Composer -->
      <div class="composer-wrapper">
        <div class="composer">
          <!-- Reference thumbnails -->
          <div v-if="referenceImages.length && !replyTarget" class="composer-refs">
            <div v-for="(ref, ri) in referenceImages" :key="ri" class="composer-ref-item">
              <button class="composer-ref-preview" @click="openRefLightbox(ri)">
                <img :src="ref.dataUrl" :alt="ref.name" />
              </button>
              <button class="composer-ref-remove" @click="removeReference(ri)" title="移除"><Icon name="x" size="xs" /></button>
            </div>
          </div>

          <!-- Reply banner -->
          <div v-if="replyTarget" class="composer-reply-banner" @click.stop>
            <span class="composer-reply-icon"><Icon name="arrowDown" size="sm" /></span>
            <div class="composer-reply-text">
              <span class="composer-reply-heading">正在回复 AI 的提问 · 无需粘贴原文，模型会自动收到上下文</span>
              <p class="composer-reply-body">{{ replyTarget.aiMessage }}</p>
            </div>
            <button class="composer-reply-cancel" @click="cancelReply"><Icon name="x" size="sm" /></button>
          </div>

          <div class="composer-inner" @click="focusTextarea">
            <textarea
              ref="textareaRef"
              v-model="prompt"
              class="composer-textarea"
              :placeholder="replyTarget ? '输入你的回答…' : referenceImages.length ? '描述你希望如何修改参考图' : '输入你想要生成的画面，也可直接粘贴图片'"
              @keydown.enter.exact.prevent="submit"
              @paste="onPaste"
              @input="autoResize"
            ></textarea>

            <div class="composer-toolbar" ref="composerControlsRef" @click.stop>
              <div class="composer-toolbar-left">
                <div class="composer-field composer-field-size">
                  <span class="composer-field-label">尺寸</span>
                  <button class="composer-field-control composer-size-control" type="button" @click="openSizeDialog">
                    <span class="font-data">{{ sizeDisplayLabel }}</span>
                  </button>
                </div>

                <div class="composer-field">
                  <span class="composer-field-label">质量</span>
                  <div class="composer-select-wrap">
                    <button class="composer-field-control" type="button" :class="{ 'composer-field-control-open': openSelectMenu === 'quality' }" @click="toggleSelectMenu('quality')">
                      <span>{{ qualityDisplayLabel }}</span>
                      <Icon name="chevronDown" size="xs" :class="{ 'rotate-180': openSelectMenu === 'quality' }" />
                    </button>
                    <div v-if="openSelectMenu === 'quality'" class="composer-select-menu">
                      <button v-for="opt in resolutionOptions" :key="opt.value" type="button" class="composer-select-option" :class="{ 'composer-select-option-active': opt.value === resolution }" @click="selectFieldValue('quality', opt.value)">
                        {{ opt.value }}
                      </button>
                    </div>
                  </div>
                </div>

                <div class="composer-field">
                  <span class="composer-field-label">输出格式</span>
                  <div class="composer-select-wrap">
                    <button class="composer-field-control" type="button" :class="{ 'composer-field-control-open': openSelectMenu === 'format' }" @click="toggleSelectMenu('format')">
                      <span>{{ outputFormatLabel }}</span>
                      <Icon name="chevronDown" size="xs" :class="{ 'rotate-180': openSelectMenu === 'format' }" />
                    </button>
                    <div v-if="openSelectMenu === 'format'" class="composer-select-menu">
                      <button v-for="opt in outputFormatOptions" :key="opt.value" type="button" class="composer-select-option" :class="{ 'composer-select-option-active': opt.value === outputFormat }" @click="selectFieldValue('format', opt.value)">
                        {{ opt.label }}
                      </button>
                    </div>
                  </div>
                </div>

                <div class="composer-field">
                  <span class="composer-field-label">透明背景</span>
                  <div class="composer-select-wrap">
                    <button class="composer-field-control" type="button" :class="{ 'composer-field-control-open': openSelectMenu === 'background' }" @click="toggleSelectMenu('background')">
                      <span>{{ backgroundLabel }}</span>
                      <Icon name="chevronDown" size="xs" :class="{ 'rotate-180': openSelectMenu === 'background' }" />
                    </button>
                    <div v-if="openSelectMenu === 'background'" class="composer-select-menu">
                      <button v-for="opt in backgroundOptions" :key="opt.value" type="button" class="composer-select-option" :class="{ 'composer-select-option-active': opt.value === background }" @click="selectFieldValue('background', opt.value)">
                        {{ opt.label }}
                      </button>
                    </div>
                  </div>
                </div>

                <div class="composer-field">
                  <span class="composer-field-label">审核</span>
                  <div class="composer-select-wrap">
                    <button class="composer-field-control" type="button" :class="{ 'composer-field-control-open': openSelectMenu === 'moderation' }" @click="toggleSelectMenu('moderation')">
                      <span>{{ moderation }}</span>
                      <Icon name="chevronDown" size="xs" :class="{ 'rotate-180': openSelectMenu === 'moderation' }" />
                    </button>
                    <div v-if="openSelectMenu === 'moderation'" class="composer-select-menu">
                      <button v-for="opt in moderationOptions" :key="opt.value" type="button" class="composer-select-option" :class="{ 'composer-select-option-active': opt.value === moderation }" @click="selectFieldValue('moderation', opt.value)">
                        {{ opt.label }}
                      </button>
                    </div>
                  </div>
                </div>

                <label class="composer-field composer-field-count">
                  <span class="composer-field-label">数量</span>
                  <input v-model="count" class="composer-count-input" type="number" min="1" :max="maxCount" inputmode="numeric" @blur="normalizeComposerOptions" />
                </label>

                <div class="composer-actions">
                  <button class="composer-icon-action" type="button" @click="pickReference" :title="referenceImages.length ? '添加参考图' : '上传参考图'">
                    <Icon name="link" size="md" />
                  </button>
                  <button class="composer-submit" :disabled="!prompt.trim()" @click="submit" :title="referenceImages.length ? '编辑图片' : '生成图片'">
                    <Icon name="arrowRight" size="md" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <input ref="fileInput" type="file" accept="image/*" multiple class="hidden" @change="onFilesSelected" />

      <!-- History Dialog -->
      <Teleport to="body">
        <div v-if="showSizeDialog" class="dialog-overlay size-dialog-overlay" @click.self="cancelSizeDialog">
          <div class="size-dialog">
            <div class="size-dialog-header">
              <div>
                <h2 class="size-dialog-title">设置图像尺寸</h2>
                <p class="size-dialog-current">当前：{{ sizeDisplayLabel }}</p>
              </div>
              <button class="size-dialog-close" type="button" @click="cancelSizeDialog"><Icon name="x" size="md" /></button>
            </div>

            <div class="size-tabs">
              <button type="button" class="size-tab" :class="{ 'size-tab-active': sizeDraftMode === 'auto' }" @click="selectSizeDraftMode('auto')">自动</button>
              <button type="button" class="size-tab" :class="{ 'size-tab-active': sizeDraftMode === 'ratio' }" @click="selectSizeDraftMode('ratio')">按比例</button>
              <button type="button" class="size-tab" :class="{ 'size-tab-active': sizeDraftMode === 'custom' }" @click="selectSizeDraftMode('custom')">自定义宽高</button>
            </div>

            <div class="size-dialog-body">
              <template v-if="sizeDraftMode === 'auto'">
                <button type="button" class="size-auto-option" :class="{ 'size-auto-option-active': sizeDraft === 'auto' }" @click="sizeDraft = 'auto'">
                  <span>自动匹配</span>
                  <small>由模型或后台默认配置决定最终尺寸</small>
                </button>
              </template>

              <template v-else-if="sizeDraftMode === 'ratio'">
                <p class="size-section-label">图像比例</p>
                <div class="size-ratio-grid">
                  <button v-for="opt in ratioSizeOptions" :key="opt.value" type="button" class="size-ratio-card" :class="{ 'size-ratio-card-active': opt.value === sizeDraft }" @click="sizeDraft = opt.value">
                    <span class="size-ratio-icon">
                      <span :style="{ width: opt.w * 0.7 + 'px', height: opt.h * 0.7 + 'px' }"></span>
                    </span>
                    <span class="font-data">{{ ratioLabel(opt.value) }}</span>
                    <small>{{ opt.label }}</small>
                    <em v-if="mappedSizeLabel(opt.value) !== opt.label">使用 {{ mappedSizeLabel(opt.value) }}</em>
                  </button>
                </div>
              </template>

              <template v-else>
                <p class="size-section-label">输入具体像素值</p>
                <div class="size-custom-row">
                  <label>
                    <span>宽度 (Width)</span>
                    <input v-model="customWidth" type="number" min="1" inputmode="numeric" />
                  </label>
                  <span class="size-custom-times">×</span>
                  <label>
                    <span>高度 (Height)</span>
                    <input v-model="customHeight" type="number" min="1" inputmode="numeric" />
                  </label>
                </div>
                <div class="size-info-box">
                  <Icon name="infoCircle" size="sm" />
                  <span>将根据后台允许尺寸自动匹配最接近的可用值。</span>
                </div>
              </template>
            </div>

            <div class="size-dialog-result">
              <span>将使用</span>
              <strong class="font-data">{{ sizeDialogUsedLabel }}</strong>
            </div>

            <div class="size-dialog-footer">
              <button type="button" class="size-dialog-secondary" @click="cancelSizeDialog">取消</button>
              <button type="button" class="size-dialog-primary" @click="confirmSizeDialog">确定</button>
            </div>
          </div>
        </div>

        <div v-if="isHistoryOpen" class="dialog-overlay" @click.self="isHistoryOpen = false">
          <div class="history-dialog">
            <div class="history-dialog-header">
              <div class="flex items-center gap-2"><Icon name="clock" size="sm" /><span class="font-semibold">历史对话</span><span class="history-count">{{ conversations.length }}</span></div>
            </div>
            <div class="history-dialog-body">
              <div v-if="loadingHistory" class="history-loading"><Icon name="refresh" size="sm" class="animate-spin" />加载中</div>
              <div v-else-if="!conversations.length" class="history-empty">还没有图片记录</div>
              <button v-for="conv in sortedConversations" :key="conv.id" class="history-item" :class="{ 'history-item-active': conv.id === activeConversationId }" @click="selectConversation(conv.id); isHistoryOpen = false">
                <div class="history-item-main">
                  <span class="history-item-title">{{ conv.title }}</span>
                  <span class="history-item-meta">{{ conv.turns.length }} 轮 · {{ formatTime(conv.updatedAt) }}</span>
                </div>
                <div class="history-item-actions" @click.stop>
                  <button class="turn-icon-btn" @click="deleteConversation(conv.id)" title="删除"><Icon name="trash" size="xs" /></button>
                </div>
              </button>
            </div>
            <div class="history-dialog-footer">
              <button class="btn btn-secondary w-full" @click="isHistoryOpen = false">关闭</button>
            </div>
          </div>
        </div>
        <div v-if="showWorkbenchHelp" class="dialog-overlay" @click.self="showWorkbenchHelp = false">
          <div class="workbench-help-dialog">
            <div class="workbench-help-header">
              <div>
                <p class="workbench-help-kicker">GPTImage</p>
                <h2>操作指南</h2>
              </div>
              <button class="workbench-help-close" type="button" title="关闭" @click="showWorkbenchHelp = false">
                <Icon name="x" size="sm" />
              </button>
            </div>
            <div class="workbench-help-body">
              <section class="workbench-help-grid">
                <article>
                  <Icon name="grid" size="sm" />
                  <strong>画廊</strong>
                  <span>当前创作工作台，保留本地历史、参考图和生成参数。</span>
                </article>
                <article>
                  <Icon name="brain" size="sm" />
                  <strong>Agent</strong>
                  <span>入口已预留，后续接入多轮规划和工具调用式生图。</span>
                </article>
                <article>
                  <Icon name="terminal" size="sm" />
                  <strong>控制台</strong>
                  <span>打开历史会话，查看运行任务数量，并复用旧任务配置。</span>
                </article>
                <article>
                  <Icon name="cog" size="sm" />
                  <strong>设置</strong>
                  <span>查看 Sub2API 图片接口、模型、请求模式和当前生成参数。</span>
                </article>
              </section>
              <section class="workbench-help-steps">
                <p>基础流程</p>
                <ol>
                  <li>在 API 密钥页创建绑定生图分组的 Key。</li>
                  <li>打开设置确认接口为当前站点的 <code>/v1</code>。</li>
                  <li>输入提示词，可粘贴或上传参考图，再提交生成。</li>
                </ol>
              </section>
            </div>
          </div>
        </div>
        <div v-if="showWorkbenchSettings" class="dialog-overlay" @click.self="showWorkbenchSettings = false">
          <div class="workbench-settings-dialog">
            <div class="workbench-help-header">
              <div>
                <p class="workbench-help-kicker">GPTImage</p>
                <h2>接口与生成设置</h2>
              </div>
              <button class="workbench-help-close" type="button" title="关闭" @click="showWorkbenchSettings = false">
                <Icon name="x" size="sm" />
              </button>
            </div>
            <div class="workbench-settings-body">
              <section class="settings-section">
                <div class="settings-section-title">
                  <Icon name="server" size="sm" />
                  <span>Sub2API 接口</span>
                </div>
                <dl class="settings-list">
                  <div><dt>API URL</dt><dd>{{ apiBaseUrl }}</dd></div>
                  <div><dt>生成端点</dt><dd>{{ imageGenerationEndpoint }}</dd></div>
                  <div><dt>编辑端点</dt><dd>{{ imageEditEndpoint }}</dd></div>
                  <div><dt>API mode</dt><dd>images</dd></div>
                  <div><dt>模型</dt><dd>{{ activeImageModel }}</dd></div>
                  <div><dt>API Key</dt><dd>使用 API 密钥页中绑定生图分组的 Key</dd></div>
                </dl>
                <router-link class="settings-link" to="/keys" @click="showWorkbenchSettings = false">
                  <Icon name="key" size="sm" />
                  管理 API 密钥
                </router-link>
              </section>
              <section class="settings-section">
                <div class="settings-section-title">
                  <Icon name="filter" size="sm" />
                  <span>当前生成参数</span>
                </div>
                <div class="settings-chip-grid">
                  <span><b>尺寸</b>{{ sizeDisplayLabel }}</span>
                  <span><b>质量</b>{{ qualityDisplayLabel }}</span>
                  <span><b>格式</b>{{ outputFormatLabel }}</span>
                  <span><b>背景</b>{{ backgroundLabel }}</span>
                  <span><b>审核</b>{{ moderation }}</span>
                  <span><b>数量</b>{{ parsedCount }}</span>
                </div>
                <button class="settings-focus-button" type="button" @click="showWorkbenchSettings = false; focusComposerSettings()">
                  <Icon name="edit" size="sm" />
                  调整生成参数
                </button>
              </section>
              <section class="settings-section settings-section-compact">
                <div class="settings-section-title">
                  <Icon name="shield" size="sm" />
                  <span>运行边界</span>
                </div>
                <p>同源使用无需 CORS；请求仍走 Sub2API 的鉴权、分组权限、计费和使用记录。</p>
              </section>
            </div>
          </div>
        </div>
        <div v-if="confirmClearAll" class="dialog-overlay" @click.self="confirmClearAll = false">
          <div class="confirm-dialog">
            <h2 class="font-semibold text-lg mb-2">清空历史记录</h2>
            <p class="text-sm text-stone-500 mb-4">确认删除全部图片历史记录吗？删除后无法恢复。</p>
            <div class="flex justify-end gap-2">
              <button class="btn btn-secondary" @click="confirmClearAll = false">取消</button>
              <button class="btn btn-danger" @click="clearHistory">确认删除</button>
            </div>
          </div>
        </div>
      </Teleport>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'

// ── Constants ──
interface ImageGenerationConfig {
  enabled: boolean
  user_workspace_enabled: boolean
  allow_edits: boolean
  default_model: string
  allowed_models: string[]
  sizes: string[]
  qualities: string[]
  output_formats: string[]
  backgrounds: string[]
  moderations: string[]
  response_formats: string[]
  default_response_format: string
  max_count: number
  request_timeout_seconds: number
}

const DEFAULT_IMAGE_CONFIG: ImageGenerationConfig = {
  enabled: true,
  user_workspace_enabled: true,
  allow_edits: true,
  default_model: 'gpt-image-2',
  allowed_models: ['gpt-image-2'],
  sizes: ['auto', '1024x1024', '1536x1024', '1024x1536'],
  qualities: ['auto', 'low', 'medium', 'high'],
  output_formats: ['png', 'jpeg', 'webp'],
  backgrounds: ['auto', 'transparent', 'opaque'],
  moderations: ['auto', 'low'],
  response_formats: ['b64_json', 'url'],
  default_response_format: 'b64_json',
  max_count: 4,
  request_timeout_seconds: 300,
}

const COMMON_RATIO_PRESET_SIZES = [
  '1024x1024',
  '1536x1024',
  '1024x1536',
  '1600x900',
  '900x1600',
  '1200x900',
  '900x1200',
  '2100x900',
]

const QUALITY_LABELS: Record<string, { label: string; desc: string }> = {
  auto: { label: '自动', desc: '由上游决定' },
  low: { label: '低', desc: '更快更省' },
  medium: { label: '中', desc: '均衡质量' },
  high: { label: '高', desc: '优先画质' },
}

const OUTPUT_FORMAT_LABELS: Record<string, string> = {
  png: 'PNG',
  jpeg: 'JPEG',
  jpg: 'JPG',
  webp: 'WebP',
}

const BACKGROUND_LABELS: Record<string, string> = {
  transparent: 'true',
  opaque: 'false',
  auto: 'auto',
}

const STORAGE_KEYS = {
  conversations: 'image_generation_conversations:v2',
  activeId: 'image_generation_active_id:v1',
  lastSize: 'image_generation_last_size',
  lastResolution: 'image_generation_last_resolution',
  lastOutputFormat: 'image_generation_last_output_format',
  lastBackground: 'image_generation_last_background',
  lastModeration: 'image_generation_last_moderation',
  lastCount: 'image_generation_last_count',
  scrollPositions: 'image_generation_scroll_positions:v1',
}

// ── Types ──
interface StoredImage {
  id: string; status: 'loading' | 'success' | 'error'
  src?: string; b64_json?: string; url?: string
  mime_type?: string; error?: string; sizeLabel?: string
}
interface StoredReference { name: string; type: string; dataUrl: string }
interface ImageTurn {
  id: string; prompt: string; mode: 'generate' | 'edit'
  referenceImages: StoredReference[]; model: string; count: number
  size: string; resolution: string
  outputFormat?: string; background?: string; moderation?: string
  images: StoredImage[]; status: 'queued' | 'generating' | 'success' | 'error'
  error?: string; createdAt: string; promptDeleted: boolean; resultsDeleted: boolean
}
interface ImageConversation { id: string; title: string; createdAt: string; updatedAt: string; turns: ImageTurn[] }
interface SizeOption { value: string; label: string; desc: string; w: number; h: number }
interface QualityOption { value: string; label: string; desc: string }
interface SimpleOption { value: string; label: string }
type SelectMenu = 'quality' | 'format' | 'background' | 'moderation'
type SizeDraftMode = 'auto' | 'ratio' | 'custom'
// ── State ──
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const resultsViewport = ref<HTMLDivElement | null>(null)
const composerControlsRef = ref<HTMLElement | null>(null)

const prompt = ref('')
const count = ref('1')
const size = ref('')
const resolution = ref('')
const outputFormat = ref('')
const background = ref('')
const moderation = ref('')
const referenceImages = ref<StoredReference[]>([])
const referenceFiles = ref<File[]>([])
const isHistoryOpen = ref(false)
const showWorkbenchHelp = ref(false)
const showWorkbenchSettings = ref(false)
const loadingHistory = ref(false)
const conversations = ref<ImageConversation[]>([])
const activeConversationId = ref<string | null>(null)
const replyTarget = ref<{ aiMessage: string } | null>(null)
const confirmClearAll = ref(false)
const showSizeDialog = ref(false)
const openSelectMenu = ref<SelectMenu | null>(null)
const sizeDraftMode = ref<SizeDraftMode>('ratio')
const sizeDraft = ref('')
const customWidth = ref('1024')
const customHeight = ref('1024')
const quotaDisplay = ref('本地')
const imageConfig = ref<ImageGenerationConfig>({ ...DEFAULT_IMAGE_CONFIG })

// ── Computed ──
const maxCount = computed(() => Math.max(1, Math.min(8, Number(imageConfig.value.max_count) || DEFAULT_IMAGE_CONFIG.max_count)))
const parsedCount = computed(() => Math.max(1, Math.min(maxCount.value, Number(count.value) || 1)))
const sizeOptions = computed(() => optionValues(imageConfig.value.sizes, DEFAULT_IMAGE_CONFIG.sizes).map(toSizeOption))
const allowedSizeOptions = computed(() => sizeOptions.value.filter(o => o.value !== 'auto'))
const ratioSizeOptions = computed(() => {
  const seen = new Set<string>()
  const result: SizeOption[] = []
  for (const value of COMMON_RATIO_PRESET_SIZES) {
    if (seen.has(value)) continue
    seen.add(value)
    result.push(toSizeOption(value))
  }
  for (const opt of allowedSizeOptions.value) {
    if (seen.has(opt.value)) continue
    seen.add(opt.value)
    result.push(opt)
  }
  return result
})
const resolutionOptions = computed(() => optionValues(imageConfig.value.qualities, DEFAULT_IMAGE_CONFIG.qualities).map(toQualityOption))
const outputFormatOptions = computed(() => optionValues(imageConfig.value.output_formats, DEFAULT_IMAGE_CONFIG.output_formats).map(toSimpleOption))
const backgroundOptions = computed(() => optionValues(imageConfig.value.backgrounds, DEFAULT_IMAGE_CONFIG.backgrounds).map(toBackgroundOption))
const moderationOptions = computed(() => optionValues(imageConfig.value.moderations, DEFAULT_IMAGE_CONFIG.moderations).map(toSimpleOption))
const defaultSizeValue = computed(() => allowedSizeOptions.value[0]?.value || sizeOptions.value[0]?.value || '1024x1024')
const defaultQualityValue = computed(() => resolutionOptions.value.find(o => o.value === 'high')?.value || resolutionOptions.value[0]?.value || DEFAULT_IMAGE_CONFIG.qualities[0])
const defaultOutputFormatValue = computed(() => outputFormatOptions.value.find(o => o.value === 'png')?.value || outputFormatOptions.value[0]?.value || DEFAULT_IMAGE_CONFIG.output_formats[0])
const defaultBackgroundValue = computed(() => backgroundOptions.value.find(o => o.value === 'opaque')?.value || backgroundOptions.value[0]?.value || DEFAULT_IMAGE_CONFIG.backgrounds[0])
const defaultModerationValue = computed(() => moderationOptions.value.find(o => o.value === 'auto')?.value || moderationOptions.value[0]?.value || DEFAULT_IMAGE_CONFIG.moderations[0])
const sizeOption = computed(() => sizeOptions.value.find(o => o.value === size.value) ?? sizeOptions.value[0])
const sizeDisplayLabel = computed(() => size.value || sizeOption.value?.label || defaultSizeValue.value)
const qualityDisplayLabel = computed(() => resolution.value || resolutionOptions.value[0]?.value || defaultQualityValue.value)
const outputFormatLabel = computed(() => outputFormatOptions.value.find(o => o.value === outputFormat.value)?.label ?? outputFormat.value)
const backgroundLabel = computed(() => backgroundOptions.value.find(o => o.value === background.value)?.label ?? background.value)
const sizeDialogUsedValue = computed(() => {
  if (sizeDraftMode.value === 'auto') return selectedOptionValue('auto', sizeOptions.value, defaultSizeValue.value)
  if (sizeDraftMode.value === 'custom') return closestSizeToDimensions(Number(customWidth.value), Number(customHeight.value))
  const parsed = parseSizeValue(sizeDraft.value)
  return parsed ? closestSizeToDimensions(parsed.width, parsed.height) : defaultSizeValue.value
})
const sizeDialogUsedLabel = computed(() => sizeOptions.value.find(o => o.value === sizeDialogUsedValue.value)?.label ?? sizeDialogUsedValue.value)
const activeConversation = computed(() => conversations.value.find(c => c.id === activeConversationId.value) ?? null)
const sortedConversations = computed(() => [...conversations.value].sort((a, b) => b.updatedAt.localeCompare(a.updatedAt)))
const apiBaseUrl = computed(() => {
  if (typeof window === 'undefined') return '/v1'
  return `${window.location.origin}/v1`
})
const imageGenerationEndpoint = computed(() => `${apiBaseUrl.value}/images/generations`)
const imageEditEndpoint = computed(() => `${apiBaseUrl.value}/images/edits`)
const activeImageModel = computed(() => imageConfig.value.default_model || 'gpt-image-2')
const activeTaskCount = computed(() => {
  let sum = 0
  for (const c of conversations.value) {
    for (const t of c.turns) {
      if (t.status === 'queued' || t.status === 'generating') sum++
    }
  }
  return sum
})

// ── Helpers ──
function uid(): string { return crypto.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(16).slice(2)}` }
function formatTime(v: string) { const d = new Date(v); return Number.isNaN(d.getTime()) ? '' : new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(d) }
function optionValues(values: string[] | undefined, fallback: string[]): string[] {
  const source = Array.isArray(values) && values.length ? values : fallback
  const seen = new Set<string>()
  const result: string[] = []
  for (const raw of source) {
    const value = String(raw || '').trim()
    if (!value || seen.has(value)) continue
    seen.add(value)
    result.push(value)
  }
  return result.length ? result : [...fallback]
}
function gcd(a: number, b: number): number { return b === 0 ? a : gcd(b, a % b) }
function sizePreviewDimensions(value: string): { w: number; h: number } {
  const match = value.match(/^(\d+)x(\d+)$/)
  if (!match) return { w: 22, h: 22 }
  const width = Number(match[1])
  const height = Number(match[2])
  if (!width || !height) return { w: 22, h: 22 }
  const ratio = width / height
  return ratio >= 1
    ? { w: 28, h: Math.max(14, 28 / ratio) }
    : { w: Math.max(14, 28 * ratio), h: 28 }
}
function parseSizeValue(value: string): { width: number; height: number } | null {
  const match = value.match(/^(\d+)x(\d+)$/)
  if (!match) return null
  const width = Number(match[1])
  const height = Number(match[2])
  return width > 0 && height > 0 ? { width, height } : null
}
function sizeDescription(value: string): string {
  const match = value.match(/^(\d+)x(\d+)$/)
  if (!match) return value === 'auto' ? '由上游决定' : '后台配置'
  const width = Number(match[1])
  const height = Number(match[2])
  const divisor = gcd(width, height)
  const ratio = `${width / divisor}:${height / divisor}`
  const direction = width === height ? '正方形' : width > height ? '横版' : '竖版'
  return `${ratio} · ${direction}`
}
function toSizeOption(value: string): SizeOption {
  const size = sizePreviewDimensions(value)
  return {
    value,
    label: value === 'auto' ? '自动' : value,
    desc: sizeDescription(value),
    ...size,
  }
}
function toQualityOption(value: string): QualityOption {
  const preset = QUALITY_LABELS[value]
  return preset ? { value, ...preset } : { value, label: value, desc: '后台配置' }
}
function toSimpleOption(value: string): SimpleOption {
  const key = value.toLowerCase()
  return { value, label: OUTPUT_FORMAT_LABELS[key] || value }
}
function toBackgroundOption(value: string): SimpleOption {
  return { value, label: BACKGROUND_LABELS[value] || value }
}
function selectedOptionValue(value: string, options: Array<{ value: string }>, fallback: string): string {
  return options.some((opt) => opt.value === value) ? value : fallback
}
function normalizeComposerOptions() {
  count.value = String(parsedCount.value)
  size.value = selectedOptionValue(size.value, sizeOptions.value, defaultSizeValue.value)
  resolution.value = selectedOptionValue(resolution.value, resolutionOptions.value, defaultQualityValue.value)
  outputFormat.value = selectedOptionValue(outputFormat.value, outputFormatOptions.value, defaultOutputFormatValue.value)
  background.value = selectedOptionValue(background.value, backgroundOptions.value, defaultBackgroundValue.value)
  moderation.value = selectedOptionValue(moderation.value, moderationOptions.value, defaultModerationValue.value)
}
function ratioLabel(value: string): string {
  const parsed = parseSizeValue(value)
  if (!parsed) return value
  const divisor = gcd(parsed.width, parsed.height)
  return `${parsed.width / divisor}:${parsed.height / divisor}`
}
function mappedSizeLabel(value: string): string {
  const parsed = parseSizeValue(value)
  const mapped = parsed ? closestSizeToDimensions(parsed.width, parsed.height) : selectedOptionValue(value, sizeOptions.value, defaultSizeValue.value)
  return sizeOptions.value.find(o => o.value === mapped)?.label ?? mapped
}
function closestSizeToDimensions(width: number, height: number): string {
  if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) {
    return defaultSizeValue.value
  }
  let best = allowedSizeOptions.value[0] || sizeOptions.value[0]
  let bestScore = Number.POSITIVE_INFINITY
  for (const opt of allowedSizeOptions.value) {
    const parsed = parseSizeValue(opt.value)
    if (!parsed) continue
    const ratioScore = Math.abs((parsed.width / parsed.height) - (width / height)) * 10000
    const sizeScore = Math.abs(parsed.width - width) + Math.abs(parsed.height - height)
    const score = ratioScore + sizeScore
    if (score < bestScore) {
      best = opt
      bestScore = score
    }
  }
  return best?.value || defaultSizeValue.value
}
function readFileAsDataUrl(file: File): Promise<StoredReference> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve({ name: file.name, type: file.type || 'image/png', dataUrl: String(reader.result || '') })
    reader.onerror = () => reject(new Error('读取失败'))
    reader.readAsDataURL(file)
  })
}
function dataUrlToFile(dataUrl: string, fileName: string, mimeType?: string) {
  const [header, content] = dataUrl.split(',', 2)
  const mt = header.match(/data:(.*?);base64/)?.[1] || mimeType || 'image/png'
  const binary = atob(content || '')
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return new File([bytes], fileName, { type: mt })
}

function persistConversations() {
  try { localStorage.setItem(STORAGE_KEYS.conversations, JSON.stringify(conversations.value)) } catch { /* noop */ }
}
function persistActiveId() {
  try { localStorage.setItem(STORAGE_KEYS.activeId, activeConversationId.value ?? '') } catch { /* noop */ }
}

function createLoadingImages(turnId: string, n: number): StoredImage[] {
  return Array.from({ length: n }, (_, i) => ({ id: `${turnId}-${i}`, status: 'loading' as const }))
}

// ── Persistence ──
function loadConversations() {
  try {
    const raw = localStorage.getItem(STORAGE_KEYS.conversations)
    if (raw) conversations.value = JSON.parse(raw)
  } catch { conversations.value = [] }
}
function loadActiveId() {
  try { activeConversationId.value = localStorage.getItem(STORAGE_KEYS.activeId) || null } catch { activeConversationId.value = null }
}
function loadPreferences() {
  count.value = localStorage.getItem(STORAGE_KEYS.lastCount) || '1'
  size.value = localStorage.getItem(STORAGE_KEYS.lastSize) || defaultSizeValue.value
  resolution.value = localStorage.getItem(STORAGE_KEYS.lastResolution) || defaultQualityValue.value
  outputFormat.value = localStorage.getItem(STORAGE_KEYS.lastOutputFormat) || defaultOutputFormatValue.value
  background.value = localStorage.getItem(STORAGE_KEYS.lastBackground) || defaultBackgroundValue.value
  moderation.value = localStorage.getItem(STORAGE_KEYS.lastModeration) || defaultModerationValue.value
  normalizeComposerOptions()
}
function loadImageConfig() {
  imageConfig.value = { ...DEFAULT_IMAGE_CONFIG }
  normalizeComposerOptions()
}

// ── Conversation actions ──
function newConversation() {
  activeConversationId.value = null
  clearComposer()
  persistActiveId()
  nextTick(() => textareaRef.value?.focus())
}
function selectConversation(id: string) {
  activeConversationId.value = id
  persistActiveId()
}
function deleteConversation(id: string) {
  conversations.value = conversations.value.filter(c => c.id !== id)
  if (activeConversationId.value === id) { activeConversationId.value = null; clearComposer() }
  persistConversations(); persistActiveId()
}
function clearHistory() {
  conversations.value = []
  activeConversationId.value = null
  clearComposer()
  persistConversations(); persistActiveId()
  confirmClearAll.value = false
}
function deleteTurnPrompt(turnId: string) {
  const conv = activeConversation.value; if (!conv) return
  updateConv({ ...conv, turns: conv.turns.filter(t => { if (t.id !== turnId) return true; t.promptDeleted = true; return !t.resultsDeleted }), updatedAt: new Date().toISOString() })
}
function deleteTurnResults(turnId: string) {
  const conv = activeConversation.value; if (!conv) return
  updateConv({ ...conv, turns: conv.turns.filter(t => { if (t.id !== turnId) return true; t.resultsDeleted = true; return !t.promptDeleted }), updatedAt: new Date().toISOString() })
}
function reuseConfig(turn: ImageTurn) {
  prompt.value = turn.prompt
  count.value = String(turn.count || 1)
  size.value = turn.size
  resolution.value = turn.resolution
  outputFormat.value = turn.outputFormat || outputFormat.value || defaultOutputFormatValue.value
  background.value = turn.background || background.value || defaultBackgroundValue.value
  moderation.value = turn.moderation || moderation.value || defaultModerationValue.value
  normalizeComposerOptions()
  referenceImages.value = [...turn.referenceImages]
  referenceFiles.value = turn.referenceImages.map(r => dataUrlToFile(r.dataUrl, r.name, r.type))
  nextTick(() => textareaRef.value?.focus())
}
function continueEdit(_turn: ImageTurn, img: StoredImage | StoredReference) {
  if ('dataUrl' in img) {
    referenceImages.value = [img]; referenceFiles.value = [dataUrlToFile(img.dataUrl, img.name, img.type)]
  } else if (img.b64_json) {
    const ref: StoredReference = { name: 'edit.png', type: 'image/png', dataUrl: `data:image/png;base64,${img.b64_json}` }
    referenceImages.value = [ref]; referenceFiles.value = [dataUrlToFile(ref.dataUrl, ref.name, ref.type)]
  } else if (img.url) {
    fetch(img.url).then(r => r.blob()).then(b => {
      const f = new File([b], 'edit.png', { type: b.type || 'image/png' })
      readFileAsDataUrl(f).then(ref => { referenceImages.value = [ref]; referenceFiles.value = [f] })
    }).catch(() => {})
  }
  prompt.value = ''
  nextTick(() => textareaRef.value?.focus())
}
function regenerateTurn(turn: ImageTurn) {
  if (!activeConversation.value) return
  const newTurnId = uid()
  const nextCount = Math.max(1, Math.min(maxCount.value, Number(turn.count) || 1))
  const newTurn: ImageTurn = {
    id: newTurnId,
    prompt: turn.prompt,
    model: turn.model,
    mode: turn.mode,
    referenceImages: [...turn.referenceImages],
    count: nextCount,
    size: selectedOptionValue(turn.size, sizeOptions.value, defaultSizeValue.value),
    resolution: selectedOptionValue(turn.resolution, resolutionOptions.value, defaultQualityValue.value),
    outputFormat: selectedOptionValue(turn.outputFormat || outputFormat.value, outputFormatOptions.value, defaultOutputFormatValue.value),
    background: selectedOptionValue(turn.background || background.value, backgroundOptions.value, defaultBackgroundValue.value),
    moderation: selectedOptionValue(turn.moderation || moderation.value, moderationOptions.value, defaultModerationValue.value),
    images: createLoadingImages(newTurnId, nextCount),
    createdAt: new Date().toISOString(),
    status: 'queued',
    promptDeleted: false,
    resultsDeleted: false,
  }
  const conv = { ...activeConversation.value, turns: [...activeConversation.value.turns, newTurn], updatedAt: new Date().toISOString() }
  updateConv(conv)
  runQueue(conv.id)
}
function retryImage(turnId: string, img: StoredImage) {
  const conv = activeConversation.value; if (!conv) return
  const turn = conv.turns.find(t => t.id === turnId); if (!turn) return
  const newImg: StoredImage = { id: `${turnId}-${uid()}`, status: 'loading' }
  const images = turn.images.map(i => i.id === img.id ? newImg : i)
  const newTurn = { ...turn, status: turn.status === 'queued' ? 'queued' as const : 'queued' as const, images }
  const newConv = { ...conv, turns: conv.turns.map(t => t.id === turnId ? newTurn : t), updatedAt: new Date().toISOString() }
  updateConv(newConv)
  runQueue(newConv.id)
}

function updateConv(conv: ImageConversation) {
  conversations.value = [conv, ...conversations.value.filter(c => c.id !== conv.id)].sort((a, b) => b.updatedAt.localeCompare(a.updatedAt))
  persistConversations()
}

// ── Composer ──
function clearComposer() { prompt.value = ''; referenceImages.value = []; referenceFiles.value = []; replyTarget.value = null }
function focusTextarea() { textareaRef.value?.focus() }
function autoResize() {
  const el = textareaRef.value; if (!el) return
  el.style.height = 'auto'; el.style.height = Math.min(Math.max(el.scrollHeight, 96), 360) + 'px'
}
function onPaste(e: ClipboardEvent) {
  const files = Array.from(e.clipboardData?.files || []).filter(f => f.type.startsWith('image/'))
  if (files.length) { e.preventDefault(); addReferenceFiles(files) }
}
function pickReference() { fileInput.value?.click() }
function removeReference(i: number) { referenceImages.value.splice(i, 1); referenceFiles.value.splice(i, 1) }
async function onFilesSelected(e: Event) {
  const input = e.target as HTMLInputElement
  const files = Array.from(input.files || []).filter(f => f.type.startsWith('image/'))
  await addReferenceFiles(files); input.value = ''
}
async function addReferenceFiles(files: File[]) {
  const previews = await Promise.all(files.map(f => readFileAsDataUrl(f)))
  referenceImages.value.push(...previews); referenceFiles.value.push(...files)
}
function cancelReply() { replyTarget.value = null; referenceImages.value = []; referenceFiles.value = [] }

// ── Controls ──
function closeSelectMenu() { openSelectMenu.value = null }
function toggleSelectMenu(menu: SelectMenu) { openSelectMenu.value = openSelectMenu.value === menu ? null : menu }
function selectFieldValue(menu: SelectMenu, value: string) {
  if (menu === 'quality') resolution.value = value
  else if (menu === 'format') outputFormat.value = value
  else if (menu === 'background') background.value = value
  else if (menu === 'moderation') moderation.value = value
  closeSelectMenu()
}
function openSizeDialog() {
  closeSelectMenu()
  sizeDraft.value = selectedOptionValue(size.value, sizeOptions.value, defaultSizeValue.value)
  sizeDraftMode.value = sizeDraft.value === 'auto' ? 'auto' : 'ratio'
  const parsed = parseSizeValue(sizeDraft.value) || parseSizeValue(defaultSizeValue.value)
  customWidth.value = String(parsed?.width || 1024)
  customHeight.value = String(parsed?.height || 1024)
  showSizeDialog.value = true
}
function selectSizeDraftMode(mode: SizeDraftMode) {
  sizeDraftMode.value = mode
  if (mode === 'auto') {
    sizeDraft.value = 'auto'
  } else if (mode === 'ratio') {
    sizeDraft.value = selectedOptionValue(sizeDraft.value, ratioSizeOptions.value, ratioSizeOptions.value[0]?.value || defaultSizeValue.value)
  }
}
function cancelSizeDialog() { showSizeDialog.value = false }
function confirmSizeDialog() {
  size.value = sizeDialogUsedValue.value
  showSizeDialog.value = false
}

// ── Submit ──
async function submit() {
  const p = prompt.value.trim(); if (!p) return
  normalizeComposerOptions()
  const effectiveMode = referenceFiles.value.length ? 'edit' : 'generate'
  const targetConv = activeConversation.value
  const convId = targetConv?.id ?? uid()
  const turnId = uid()
  const n = parsedCount.value
  const turnSize = selectedOptionValue(size.value, sizeOptions.value, defaultSizeValue.value)
  const turnResolution = selectedOptionValue(resolution.value, resolutionOptions.value, defaultQualityValue.value)
  const turnOutputFormat = selectedOptionValue(outputFormat.value, outputFormatOptions.value, defaultOutputFormatValue.value)
  const turnBackground = selectedOptionValue(background.value, backgroundOptions.value, defaultBackgroundValue.value)
  const turnModeration = selectedOptionValue(moderation.value, moderationOptions.value, defaultModerationValue.value)
  const turn: ImageTurn = {
    id: turnId, prompt: p, model: 'gpt-image-2', mode: effectiveMode,
    referenceImages: effectiveMode === 'edit' ? [...referenceImages.value] : [],
    count: n, size: turnSize, resolution: turnResolution,
    outputFormat: turnOutputFormat, background: turnBackground, moderation: turnModeration,
    images: createLoadingImages(turnId, n),
    createdAt: new Date().toISOString(), status: 'queued',
    promptDeleted: false, resultsDeleted: false,
  }
  const now = new Date().toISOString()
  const conv: ImageConversation = targetConv
    ? { ...targetConv, updatedAt: now, turns: [...targetConv.turns, turn] }
    : { id: convId, title: p.slice(0, 12) + (p.length > 12 ? '...' : ''), createdAt: now, updatedAt: now, turns: [turn] }
  activeConversationId.value = convId
  clearComposer()
  updateConv(conv)
  persistActiveId()
  await runQueue(convId)
}

// ── Queue runner ──
const runningQueues = new Set<string>()
async function runQueue(convId: string) {
  if (runningQueues.has(convId)) return
  runningQueues.add(convId)
  try {
    const conv = conversations.value.find(c => c.id === convId)
    const turn = conv?.turns.find(t => (t.status === 'queued' || t.status === 'generating') && t.images.some(i => i.status === 'loading'))
    if (!conv || !turn) return
    // Mark as generating
    const generatingTurn = { ...turn, status: 'generating' as const, error: undefined }
    const genConv: ImageConversation = { ...conv, turns: conv.turns.map(t => t.id === turn.id ? generatingTurn : t), updatedAt: new Date().toISOString() }
    updateConv(genConv)

    await new Promise(resolve => window.setTimeout(resolve, 250))
    const localOnlyMessage = '已保留前端画图界面，后端生图接口未接入'

    const latestConv = conversations.value.find(c => c.id === convId) ?? conv
    const images = turn.images.map((img) => {
      if (img.status !== 'loading') return img
      return { ...img, status: 'error' as const, error: localOnlyMessage }
    })
    const hasLoading = images.some(i => (i as StoredImage).status === 'loading')
    const hasSuccess = images.some(i => i.status === 'success')
    const hasError = images.some(i => i.status === 'error')
    const finalStatus: ImageTurn['status'] = hasLoading ? 'generating' : hasError ? 'error' : 'success'
    const finalTurn: ImageTurn = { ...generatingTurn, images, status: finalStatus, error: hasError && !hasSuccess ? localOnlyMessage : undefined }
    const finalConv: ImageConversation = { ...latestConv, turns: latestConv.turns.map(t => t.id === turn.id ? finalTurn : t), updatedAt: new Date().toISOString() }
    updateConv(finalConv)
  } catch (e) { /* handled per-turn */ } finally { runningQueues.delete(convId) }
}

// ── Lightbox ──
function openLightbox(turn: ImageTurn, index: number) {
  const images = turn.images.filter(i => i.status === 'success' && i.src).map(i => ({ id: i.id, src: i.src! }))
  if (images.length) window.open(images[Math.min(index, images.length - 1)].src, '_blank', 'noopener,noreferrer')
}
function openRefLightbox(index: number) {
  const ref = referenceImages.value[index]; if (ref) window.open(ref.dataUrl, '_blank', 'noopener,noreferrer')
}
async function downloadImage(img: StoredImage, index: number) {
  let blob: Blob
  if (img.b64_json) { const binary = atob(img.b64_json); const bytes = new Uint8Array(binary.length); for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i); blob = new Blob([bytes], { type: 'image/png' }) }
  else if (img.url) { const r = await fetch(img.url); blob = await r.blob() }
  else return
  const url = URL.createObjectURL(blob); const a = document.createElement('a'); a.href = url; a.download = `image-${index + 1}.png`; document.body.appendChild(a); a.click(); document.body.removeChild(a); URL.revokeObjectURL(url)
}
// ── Lifecycle ──
onMounted(async () => {
  loadConversations(); loadActiveId()
  loadImageConfig()
  loadPreferences()
  document.addEventListener('click', onDocClick)
})
onBeforeUnmount(() => { document.removeEventListener('click', onDocClick) })

function onDocClick(e: MouseEvent) {
  const t = e.target as HTMLElement
  if (!composerControlsRef.value?.contains(t)) closeSelectMenu()
}

async function focusComposer() {
  await nextTick()
  textareaRef.value?.focus()
}

async function focusComposerSettings() {
  await nextTick()
  composerControlsRef.value?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  window.setTimeout(() => {
    composerControlsRef.value?.querySelector<HTMLElement>('button, input')?.focus()
  }, 180)
}

watch(count, v => { try { localStorage.setItem(STORAGE_KEYS.lastCount, v) } catch { /* noop */ } })
watch(size, v => { try { localStorage.setItem(STORAGE_KEYS.lastSize, v) } catch { /* noop */ } })
watch(resolution, v => { try { localStorage.setItem(STORAGE_KEYS.lastResolution, v) } catch { /* noop */ } })
watch(outputFormat, v => { try { localStorage.setItem(STORAGE_KEYS.lastOutputFormat, v) } catch { /* noop */ } })
watch(background, v => { try { localStorage.setItem(STORAGE_KEYS.lastBackground, v) } catch { /* noop */ } })
watch(moderation, v => { try { localStorage.setItem(STORAGE_KEYS.lastModeration, v) } catch { /* noop */ } })
</script>

<style scoped>
.image-workbench { display: flex; flex-direction: column; height: calc(100dvh - 64px - 2rem); overflow: hidden; }
/* ── Topbar ── */
.floating-topbar { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; padding: 0.625rem 0.75rem; position: relative; z-index: 20; }
.topbar-left, .topbar-right { display: flex; align-items: center; gap: 0.5rem; min-width: 0; }
.topbar-right { justify-content: flex-end; }
.topbar-btn { display: inline-flex; align-items: center; gap: 0.375rem; height: 2.25rem; padding: 0 0.75rem; border-radius: 0.5rem; font-size: 0.8125rem; font-weight: 500; color: #444; background: white; border: 1px solid #e5e7eb; cursor: pointer; transition: all .15s; white-space: nowrap; box-shadow: 0 1px 2px rgba(0,0,0,.04); }
.topbar-btn:hover { border-color: #d1d5db; background: #f9fafb; }
.topbar-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.topbar-btn-primary { background: #1e293b; color: white; border-color: #1e293b; }
.topbar-btn-primary:hover { background: #334155; }
.topbar-badge { font-size: 0.625rem; font-weight: 600; color: #9ca3af; background: #f3f4f6; border-radius: 999px; padding: 0.0625rem 0.375rem; min-width: 1.125rem; text-align: center; }
.workbench-mode-switch { display: inline-flex; align-items: center; gap: 0.125rem; height: 2.25rem; padding: 0.1875rem; border: 1px solid #e5e7eb; border-radius: 0.75rem; background: rgba(255,255,255,.78); box-shadow: 0 1px 2px rgba(15,23,42,.04); }
.workbench-mode-tab { display: inline-flex; align-items: center; justify-content: center; height: 1.75rem; min-width: 3.25rem; padding: 0 0.75rem; border: 0; border-radius: 0.5625rem; background: transparent; color: #6b7280; font-size: 0.8125rem; font-weight: 600; cursor: pointer; transition: background .15s, color .15s, box-shadow .15s; }
.workbench-mode-tab:hover:not(:disabled) { color: #111827; background: rgba(255,255,255,.72); }
.workbench-mode-tab-active { background: #fff; color: #111827; box-shadow: 0 1px 4px rgba(15,23,42,.08); }
.workbench-mode-tab:disabled { cursor: not-allowed; opacity: .5; }
.workbench-control-strip { display: inline-flex; align-items: center; gap: 0.25rem; height: 2.25rem; padding: 0.1875rem; border: 1px solid #e5e7eb; border-radius: 0.875rem; background: rgba(255,255,255,.84); box-shadow: 0 1px 2px rgba(15,23,42,.04); }
.workbench-status-chip { display: inline-flex; align-items: center; gap: 0.375rem; height: 1.75rem; padding: 0 0.625rem; border-radius: 0.625rem; color: #64748b; background: #f8fafc; font-size: 0.75rem; font-weight: 700; white-space: nowrap; }
.workbench-status-dot { width: 0.375rem; height: 0.375rem; border-radius: 999px; background: #10b981; box-shadow: 0 0 0 3px rgba(16,185,129,.12); }
.workbench-console-btn, .workbench-icon-btn, .workbench-works-link { display: inline-flex; align-items: center; justify-content: center; gap: 0.375rem; height: 1.875rem; border: 0; border-radius: 0.625rem; background: transparent; color: #64748b; font-size: 0.8125rem; font-weight: 600; cursor: pointer; text-decoration: none; transition: background .15s, color .15s; white-space: nowrap; }
.workbench-console-btn { padding: 0 0.625rem; }
.workbench-icon-btn { width: 1.875rem; padding: 0; }
.workbench-works-link { height: 2.25rem; padding: 0 0.75rem; border: 1px solid #e5e7eb; background: rgba(255,255,255,.84); box-shadow: 0 1px 2px rgba(15,23,42,.04); }
.workbench-console-btn:hover, .workbench-icon-btn:hover, .workbench-works-link:hover { background: #f1f5f9; color: #0f172a; }
.workbench-count { display: inline-flex; min-width: 1.125rem; height: 1.125rem; align-items: center; justify-content: center; border-radius: 999px; padding: 0 0.3125rem; background: #eef2ff; color: #4f46e5; font-size: 0.6875rem; font-weight: 800; font-variant-numeric: tabular-nums; }
.dark .topbar-btn { background: #1e293b; border-color: #334155; color: #d1d5db; }
.dark .topbar-btn:hover { background: #334155; }
.dark .topbar-btn-primary { background: white; color: #0f172a; border-color: white; }
.dark .workbench-mode-switch, .dark .workbench-control-strip, .dark .workbench-works-link { background: rgba(30,41,59,.82); border-color: #334155; box-shadow: none; }
.dark .workbench-mode-tab { color: #94a3b8; }
.dark .workbench-mode-tab:hover:not(:disabled) { color: #f8fafc; background: rgba(255,255,255,.06); }
.dark .workbench-mode-tab-active { background: rgba(255,255,255,.12); color: #fff; box-shadow: none; }
.dark .workbench-status-chip { background: rgba(15,23,42,.7); color: #cbd5e1; }
.dark .workbench-console-btn, .dark .workbench-icon-btn, .dark .workbench-works-link { color: #cbd5e1; }
.dark .workbench-console-btn:hover, .dark .workbench-icon-btn:hover, .dark .workbench-works-link:hover { background: rgba(255,255,255,.08); color: #fff; }
.dark .workbench-count { background: rgba(99,102,241,.2); color: #c4b5fd; }
.topbar-btn-label { max-width: 120px; overflow: hidden; text-overflow: ellipsis; }
@media (max-width: 920px) {
  .floating-topbar { align-items: flex-start; flex-wrap: wrap; }
  .topbar-right { width: 100%; justify-content: flex-start; overflow-x: auto; padding-bottom: 0.125rem; }
}
@media (max-width: 640px) {
  .topbar-btn-label, .workbench-status-chip, .workbench-console-label, .workbench-works-label { display: none; }
  .workbench-mode-tab { min-width: auto; padding: 0 0.625rem; }
  .workbench-console-btn { width: 1.875rem; padding: 0; }
}

/* ── Results ── */
.results-viewport { flex: 1; min-height: 0; overflow-y: auto; padding: 1rem 0.5rem 0; }
.results-viewport-empty { display: flex; align-items: center; justify-content: center; }
.empty-hero { text-align: center; max-width: 40rem; padding: 2rem; position: relative; z-index: 0; }
.empty-hero-eyebrow { display: flex; align-items: center; justify-content: center; gap: 0.75rem; margin-bottom: 1.25rem; font-size: 0.625rem; font-weight: 600; letter-spacing: 0.32em; color: #78716c; text-transform: uppercase; }
.empty-hero-eyebrow span:nth-child(odd) { display: block; height: 1px; width: 2.5rem; background: #d6d3d1; }
.empty-hero-title { font-size: 2.25rem; font-weight: 600; letter-spacing: -0.02em; color: #1c1917; font-family: 'Palatino Linotype', 'Book Antiqua', serif; }
.dark .empty-hero-title { color: #fafaf9; }
.empty-hero-sub { margin-top: 0.75rem; font-size: 0.9375rem; font-style: italic; color: #78716c; max-width: 20rem; margin-inline: auto; }
.empty-hero-index { display: flex; align-items: center; justify-content: center; gap: 0.75rem; margin-top: 2rem; font-size: 0.625rem; letter-spacing: 0.2em; color: #a8a29e; }
.empty-hero-index span:nth-child(even) { display: block; height: 1px; width: 2.5rem; background: #d6d3d1; }

/* ── Turns ── */
.conversation-turns { max-width: 980px; margin: 0 auto; display: flex; flex-direction: column; gap: 2rem; }
.turn { display: flex; flex-direction: column; gap: 1rem; }
.turn-prompt-row { display: flex; justify-content: flex-end; }
.turn-prompt-bubble { max-width: 78%; background: white; border: 1px solid #e7e5e4; border-radius: 1.375rem 1.375rem 0.25rem 1.375rem; padding: 0.75rem 1rem; box-shadow: 0 1px 2px rgba(0,0,0,.04); }
.dark .turn-prompt-bubble { background: #1c1917; border-color: #292524; }
.turn-prompt-meta { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 0.5rem; margin-bottom: 0.375rem; font-size: 0.6875rem; color: #a8a29e; }
.turn-prompt-text { white-space: pre-wrap; word-break: break-word; font-size: 0.9375rem; line-height: 1.7; color: #1c1917; }
.dark .turn-prompt-text { color: #fafaf9; }
.turn-prompt-actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 0.375rem; margin-top: 0.5rem; opacity: 0.7; }
.turn-results-row { display: flex; justify-content: flex-start; }
.turn-results-block { width: 100%; padding: 0.25rem; }
.turn-ref-row { margin-bottom: 1rem; }
.turn-ref-label { display: block; margin-bottom: 0.5rem; font-size: 0.6875rem; font-weight: 500; color: #a8a29e; }
.turn-ref-images { display: flex; flex-wrap: wrap; gap: 0.75rem; }
.turn-ref-item { display: flex; flex-direction: column; align-items: flex-start; gap: 0.375rem; }
.turn-ref-img { width: 5rem; height: 5rem; border-radius: 0.75rem; object-fit: cover; border: 1px solid #e7e5e4; cursor: pointer; }
.turn-ref-edit-btn { display: inline-flex; align-items: center; gap: 0.25rem; border-radius: 999px; background: #f5f5f4; padding: 0.25rem 0.625rem; font-size: 0.6875rem; font-weight: 500; color: #57534e; border: none; cursor: pointer; }
.turn-ref-edit-btn:hover { background: #e7e5e4; color: #1c1917; }
.turn-image-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.75rem; }
.turn-image-card { break-inside: avoid; }
.turn-image-preview { width: 100%; aspect-ratio: 1; object-fit: cover; border-radius: 0.75rem; cursor: pointer; display: block; }
.turn-image-footer { display: flex; align-items: center; gap: 0.5rem; padding: 0.375rem 0.25rem; font-size: 0.625rem; color: #a8a29e; }
.turn-image-label { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.turn-image-meta { flex-shrink: 0; }
.turn-image-actions { display: flex; gap: 0.375rem; flex-shrink: 0; }
.turn-image-loader { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; min-height: 12rem; border-radius: 0.75rem; background: #f5f5f4; color: #a8a29e; gap: 0.5rem; font-size: 0.8125rem; }
.turn-image-error-card { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 1rem; background: #fafaf9; border: 1px solid #e7e5e4; border-radius: 0.75rem; text-align: center; }
.turn-error-msg { font-size: 0.75rem; line-height: 1.5; color: #57534e; }
.turn-error-actions { display: flex; gap: 0.375rem; }
.turn-chip, .turn-chip-primary { display: inline-flex; align-items: center; gap: 0.25rem; border-radius: 999px; padding: 0.25rem 0.625rem; font-size: 0.6875rem; font-weight: 500; border: none; cursor: pointer; background: #f5f5f4; color: #57534e; }
.turn-chip:hover { background: #e7e5e4; }
.turn-chip-primary { background: #1c1917; color: white; }
.turn-chip-primary:hover { background: #292524; }
.turn-icon-btn { display: inline-flex; align-items: center; justify-content: center; width: 1.5rem; height: 1.5rem; border-radius: 999px; border: none; background: transparent; color: #a8a29e; cursor: pointer; }
.turn-icon-btn:hover { background: #f5f5f4; color: #1c1917; }
.turn-error-delete-btn:hover { background: #ffe4e6; color: #e11d48; }
.dark .turn-error-delete-btn:hover { background: rgba(225, 29, 72, .16); color: #fb7185; }
.turn-footer { display: flex; align-items: center; gap: 0.5rem; margin-top: 0.75rem; }
.turn-error-banner { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.375rem 0.75rem; border-radius: 999px; background: #f5f5f4; font-size: 0.75rem; color: #a8a29e; margin-top: 0.75rem; }

/* ── Composer ── */
.composer-wrapper { flex-shrink: 0; padding: 0 0.5rem calc(0.5rem + env(safe-area-inset-bottom)); }
.composer { max-width: 820px; margin: 0 auto; position: relative; }
.composer-refs { display: flex; gap: 0.5rem; padding: 0 0.25rem 0.5rem; flex-wrap: wrap; position: absolute; bottom: 100%; left: 0; right: 0; z-index: 10; pointer-events: none; }
.composer-refs > * { pointer-events: auto; }
.composer-ref-item { position: relative; width: 3.5rem; height: 3.5rem; flex-shrink: 0; }
.composer-ref-preview { width: 100%; height: 100%; border-radius: 0.75rem; overflow: hidden; border: 1px solid #e7e5e4; cursor: pointer; display: block; padding: 0; background: none; }
.composer-ref-preview img { width: 100%; height: 100%; object-fit: cover; }
.composer-ref-remove { position: absolute; top: -0.25rem; right: -0.25rem; width: 1.25rem; height: 1.25rem; border-radius: 999px; background: white; border: 1px solid #e7e5e4; display: grid; place-items: center; cursor: pointer; color: #78716c; }
.composer-ref-remove:hover { border-color: #d6d3d1; color: #292524; }
.composer-reply-banner { display: flex; align-items: flex-start; gap: 0.5rem; margin: 0.75rem 1rem 0; padding: 0.5rem 0.75rem; border-radius: 0.75rem; background: #fafaf9; border: 1px solid #e7e5e4; }
.composer-reply-icon { flex-shrink: 0; width: 1.25rem; height: 1.25rem; border-radius: 999px; background: white; border: 1px solid #e7e5e4; display: grid; place-items: center; color: #78716c; font-size: 0.75rem; }
.composer-reply-text { flex: 1; min-width: 0; }
.composer-reply-heading { font-size: 0.6875rem; font-weight: 500; color: #78716c; }
.composer-reply-body { font-size: 0.75rem; line-height: 1.25rem; color: #57534e; overflow: hidden; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; }
.composer-reply-cancel { flex-shrink: 0; width: 1.5rem; height: 1.5rem; border-radius: 999px; border: none; background: none; color: #a8a29e; cursor: pointer; display: grid; place-items: center; }
.composer-reply-cancel:hover { background: #e7e5e4; color: #44403c; }

.composer-inner { border-radius: 1.5rem; background: white; border: 1px solid #e7e5e4; box-shadow: 0 1px 2px rgba(0,0,0,.03), 0 12px 40px rgba(15,23,42,.08); overflow: visible; cursor: text; }
.dark .composer-inner { background: #1c1917; }
.composer-textarea { width: 100%; display: block; border: 1px solid #ececec; background: transparent; resize: none; padding: 0.75rem 1rem; margin: 1rem 1rem 0; width: calc(100% - 2rem); border-radius: 1rem; font-size: 0.875rem; line-height: 1.5; color: #1c1917; outline: none; min-height: 3rem; max-height: 8rem; font-family: inherit; }
.dark .composer-textarea { color: #fafaf9; }
.composer-textarea::placeholder { color: #a8a29e; }
.composer-toolbar { padding: 0.75rem 1rem 1rem; overflow: visible; }
.dark .composer-toolbar { border-color: #292524; }
.composer-toolbar-left { display: grid; min-width: 0; grid-template-columns: minmax(7.25rem, 1.1fr) minmax(5.75rem, .75fr) minmax(5.75rem, .75fr) minmax(6.5rem, .9fr) minmax(5.5rem, .7fr) minmax(4.75rem, .55fr) auto; gap: 0.5rem; align-items: end; overflow: visible; }
.composer-field { position: relative; min-width: 0; }
.composer-field-label { display: block; margin: 0 0 0.25rem 0.125rem; font-size: 0.6875rem; font-weight: 600; color: #8a8f98; }
.composer-field-control, .composer-count-input { display: flex; align-items: center; justify-content: space-between; gap: 0.375rem; width: 100%; height: 2.125rem; padding: 0 0.75rem; border: 1px solid #e7e7e7; border-radius: 0.75rem; background: #fff; color: #27272a; font-size: 0.8125rem; line-height: 1; box-shadow: 0 1px 2px rgba(15,23,42,.03); outline: none; transition: border-color .15s, box-shadow .15s, background .15s; }
.composer-field-control { cursor: pointer; }
.composer-field-control:hover, .composer-field-control-open, .composer-count-input:focus { border-color: #d4d4d8; box-shadow: 0 2px 8px rgba(15,23,42,.06); }
.composer-size-control { justify-content: flex-start; }
.composer-select-wrap { position: relative; }
.composer-select-menu { position: absolute; left: 0; bottom: calc(100% + 0.5rem); z-index: 90; width: 7.25rem; overflow: hidden; border: 1px solid #dde3ea; border-radius: 0.75rem; background: white; box-shadow: 0 16px 40px rgba(15,23,42,.16); }
.composer-select-option { display: flex; align-items: center; width: 100%; min-height: 2.25rem; padding: 0 0.75rem; border: none; background: white; color: #4b5563; font-size: 0.8125rem; text-align: left; cursor: pointer; }
.composer-select-option:hover, .composer-select-option-active { background: #eaf3ff; color: #2563eb; }
.composer-field-count { display: block; }
.composer-count-input { display: block; font-variant-numeric: tabular-nums; }
.composer-count-input::-webkit-outer-spin-button, .composer-count-input::-webkit-inner-spin-button { margin: 0; }
.font-data { font-variant-numeric: tabular-nums; }
.composer-actions { display: flex; align-items: center; gap: 0.375rem; min-width: max-content; }
.composer-icon-action, .composer-submit { width: 2.125rem; height: 2.125rem; border-radius: 0.75rem; border: none; cursor: pointer; display: grid; place-items: center; flex-shrink: 0; transition: background .15s, color .15s, transform .15s; }
.composer-icon-action { background: #e5e7eb; color: #4b5563; }
.composer-icon-action:hover { background: #dfe3e8; color: #111827; }
.composer-submit { background: #3b82f6; color: white; box-shadow: 0 8px 18px rgba(59,130,246,.24); }
.composer-submit:hover { background: #2563eb; }
.composer-submit:disabled { background: #e5e7eb; color: #ffffff; box-shadow: none; cursor: not-allowed; }
.dark .composer-field-control, .dark .composer-count-input, .dark .composer-select-menu, .dark .composer-select-option { background: #292524; border-color: #44403c; color: #e7e5e4; }
.dark .composer-select-option:hover, .dark .composer-select-option-active { background: #1e3a5f; color: #93c5fd; }

/* ── Size Dialog ── */
.dialog-overlay.size-dialog-overlay { background: rgba(15,23,42,.34); backdrop-filter: blur(6px); }
.size-dialog { width: min(28rem, 92vw); max-height: min(42rem, 88dvh); background: #fff; border-radius: 1.5rem; box-shadow: 0 24px 80px rgba(15,23,42,.24); display: flex; flex-direction: column; overflow: hidden; }
.size-dialog-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; padding: 1.5rem 1.375rem 0.75rem; }
.size-dialog-title { font-size: 1rem; font-weight: 700; color: #18181b; }
.size-dialog-current { margin-top: 0.375rem; font-size: 0.75rem; color: #a1a1aa; }
.size-dialog-close { width: 2rem; height: 2rem; border: none; background: transparent; color: #a1a1aa; cursor: pointer; display: grid; place-items: center; border-radius: 0.5rem; }
.size-dialog-close:hover { background: #f4f4f5; color: #52525b; }
.size-tabs { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.125rem; margin: 0 1.375rem; padding: 0.25rem; border-radius: 0.75rem; background: #f7f7f8; }
.size-tab { height: 2.125rem; border: none; border-radius: 0.625rem; background: transparent; color: #52525b; font-size: 0.8125rem; font-weight: 600; cursor: pointer; transition: background .15s, box-shadow .15s, color .15s; }
.size-tab-active { background: #fff; color: #18181b; box-shadow: 0 1px 4px rgba(15,23,42,.08); }
.size-dialog-body { flex: 1; overflow-y: auto; padding: 1.5rem 1.375rem 1rem; }
.size-section-label { margin-bottom: 0.75rem; font-size: 0.75rem; font-weight: 700; color: #a1a1aa; }
.size-auto-option { display: flex; flex-direction: column; align-items: flex-start; gap: 0.375rem; width: 100%; padding: 1rem; border: 1px solid #e5e7eb; border-radius: 0.875rem; background: #fff; color: #27272a; cursor: pointer; text-align: left; }
.size-auto-option small { color: #8a8f98; }
.size-auto-option-active { border-color: #3b82f6; background: #eff6ff; color: #2563eb; }
.size-ratio-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.5rem; }
.size-ratio-card { min-height: 4.5rem; border: 1px solid #e5e7eb; border-radius: 0.75rem; background: #fff; color: #52525b; cursor: pointer; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.1875rem; font-size: 0.75rem; transition: border-color .15s, background .15s, color .15s; }
.size-ratio-card small { color: #a1a1aa; font-size: 0.625rem; }
.size-ratio-card em { color: #94a3b8; font-size: 0.5625rem; font-style: normal; line-height: 1; }
.size-ratio-card-active { border-color: #3b82f6; background: #eff6ff; color: #2563eb; }
.size-ratio-card-active em { color: #3b82f6; }
.size-ratio-icon { height: 1.25rem; display: grid; place-items: center; }
.size-ratio-icon span { display: block; border: 1.5px solid currentColor; border-radius: 0.1875rem; }
.size-custom-row { display: grid; grid-template-columns: 1fr auto 1fr; gap: 1rem; align-items: end; }
.size-custom-row label { display: flex; flex-direction: column; gap: 0.5rem; font-size: 0.75rem; font-weight: 600; color: #71717a; }
.size-custom-row input { height: 2.375rem; border: 1px solid #e5e7eb; border-radius: 0.75rem; padding: 0 0.875rem; font-size: 0.875rem; outline: none; }
.size-custom-row input:focus { border-color: #93c5fd; box-shadow: 0 0 0 3px rgba(59,130,246,.12); }
.size-custom-times { padding-bottom: 0.625rem; color: #d4d4d8; font-size: 1.25rem; }
.size-info-box { display: flex; gap: 0.625rem; margin-top: 1.25rem; padding: 0.875rem; border: 1px solid #e5e7eb; border-radius: 0.875rem; color: #52525b; font-size: 0.75rem; line-height: 1.6; }
.size-info-box svg { color: #3b82f6; flex-shrink: 0; margin-top: 0.125rem; }
.size-dialog-result { padding: 0 2.375rem 1.25rem; }
.size-dialog-result span { display: block; font-size: 0.75rem; color: #a1a1aa; margin-bottom: 0.25rem; }
.size-dialog-result strong { font-size: 1.0625rem; color: #27272a; }
.size-dialog-footer { display: grid; grid-template-columns: 1fr 1fr; gap: 0.625rem; padding: 0 1.375rem 1.375rem; }
.size-dialog-secondary, .size-dialog-primary { height: 2.5rem; border: none; border-radius: 0.75rem; font-size: 0.875rem; font-weight: 700; cursor: pointer; }
.size-dialog-secondary { background: #f4f4f5; color: #71717a; }
.size-dialog-primary { background: #3b82f6; color: white; }
.size-dialog-primary:hover { background: #2563eb; }

/* ── Dialogs ── */
.dialog-overlay { position: fixed; inset: 0; z-index: 50; display: grid; place-items: center; background: rgba(15,23,42,.58); padding: 1rem; }
.history-dialog { width: min(440px, 92vw); max-height: min(720px, 82dvh); background: white; border-radius: 1.5rem; display: flex; flex-direction: column; overflow: hidden; }
.dark .history-dialog { background: #1c1917; border: 1px solid #292524; }
.history-dialog-header { display: flex; align-items: center; justify-content: space-between; padding: 1rem 1.5rem; border-bottom: 1px solid #f5f5f4; flex-shrink: 0; }
.dark .history-dialog-header { border-color: #292524; }
.history-dialog-body { flex: 1; overflow-y: auto; padding: 0.75rem 1rem; display: flex; flex-direction: column; gap: 0.25rem; }
.history-dialog-footer { padding: 0.75rem 1rem; border-top: 1px solid #f5f5f4; flex-shrink: 0; }
.dark .history-dialog-footer { border-color: #292524; }
.history-count { font-size: 0.6875rem; font-weight: 500; color: #a8a29e; }
.history-loading, .history-empty { display: flex; align-items: center; gap: 0.5rem; padding: 0.75rem 0.5rem; font-size: 0.875rem; color: #78716c; }
.history-item { display: flex; align-items: center; justify-content: space-between; width: 100%; text-align: left; padding: 0.75rem 1rem; border-radius: 0.5rem; border: none; background: none; cursor: pointer; border-left: 2px solid transparent; transition: all .15s; }
.history-item:hover { background: #fafaf9; }
.history-item-active { border-left-color: #1c1917; background: rgba(0,0,0,.035); }
.dark .history-item:hover { background: #292524; }
.dark .history-item-active { border-left-color: white; background: rgba(255,255,255,.05); }
.history-item-title { font-weight: 600; font-size: 0.875rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 240px; display: block; color: #1c1917; }
.dark .history-item-title { color: #fafaf9; }
.history-item-meta { font-size: 0.75rem; color: #a8a29e; margin-top: 0.25rem; }
.history-item-actions { display: flex; gap: 0.125rem; flex-shrink: 0; }

.key-dialog { width: min(28rem, 92vw); max-height: min(32rem, 80dvh); background: white; border-radius: 1.5rem; display: flex; flex-direction: column; overflow: hidden; }
.dark .key-dialog { background: #1c1917; border: 1px solid #292524; }
.key-option { display: flex; align-items: center; gap: 0.75rem; width: 100%; text-align: left; padding: 0.75rem; border-radius: 0.5rem; border: 1px solid #e7e5e4; background: white; cursor: pointer; transition: all .15s; margin-bottom: 0.25rem; }
.key-option:hover { border-color: #60a5fa; }
.key-option-selected { border-color: #3b82f6; background: #eff6ff; }
.dark .key-option { border-color: #44403c; background: #292524; }
.dark .key-option-selected { border-color: #60a5fa; background: #1e3a5f; }

.workbench-help-dialog, .workbench-settings-dialog { width: min(42rem, 92vw); max-height: min(42rem, 88dvh); background: white; border-radius: 1.25rem; box-shadow: 0 24px 80px rgba(15,23,42,.22); overflow: hidden; display: flex; flex-direction: column; }
.workbench-settings-dialog { width: min(48rem, 92vw); }
.workbench-help-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; padding: 1.25rem 1.25rem 0.875rem; border-bottom: 1px solid #f1f5f9; }
.workbench-help-kicker { margin-bottom: 0.25rem; color: #0f766e; font-size: 0.75rem; font-weight: 800; }
.workbench-help-header h2 { color: #0f172a; font-size: 1rem; font-weight: 800; }
.workbench-help-close { width: 2rem; height: 2rem; border: 0; border-radius: 0.625rem; background: transparent; color: #94a3b8; cursor: pointer; display: grid; place-items: center; }
.workbench-help-close:hover { background: #f1f5f9; color: #0f172a; }
.workbench-help-body, .workbench-settings-body { display: grid; gap: 1rem; overflow-y: auto; padding: 1rem 1.25rem 1.25rem; }
.workbench-help-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.75rem; }
.workbench-help-grid article { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 0.25rem 0.625rem; align-items: start; min-width: 0; border: 1px solid #e5e7eb; border-radius: 0.875rem; background: #f8fafc; padding: 0.875rem; }
.workbench-help-grid svg { grid-row: span 2; margin-top: 0.125rem; color: #0f766e; }
.workbench-help-body strong { color: #0f172a; font-size: 0.8125rem; }
.workbench-help-body span { color: #64748b; font-size: 0.8125rem; line-height: 1.6; }
.workbench-help-steps { border: 1px solid #e5e7eb; border-radius: 0.875rem; padding: 0.875rem 1rem; }
.workbench-help-steps p, .settings-section-title { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.625rem; color: #0f172a; font-size: 0.8125rem; font-weight: 800; }
.workbench-help-steps ol { margin: 0; padding-left: 1.125rem; color: #64748b; font-size: 0.8125rem; line-height: 1.75; }
.workbench-help-steps code { border-radius: 0.25rem; background: #eef2ff; padding: 0.0625rem 0.25rem; color: #4f46e5; }
.settings-section { display: grid; gap: 0.75rem; border: 1px solid #e5e7eb; border-radius: 0.875rem; background: #fff; padding: 0.875rem; }
.settings-section-compact p { color: #64748b; font-size: 0.8125rem; line-height: 1.65; }
.settings-section-title { margin-bottom: 0; }
.settings-section-title svg { color: #0f766e; }
.settings-list { display: grid; gap: 0.5rem; margin: 0; }
.settings-list div { display: grid; grid-template-columns: 6.5rem minmax(0, 1fr); gap: 0.75rem; align-items: start; border-radius: 0.625rem; background: #f8fafc; padding: 0.625rem 0.75rem; }
.settings-list dt { color: #64748b; font-size: 0.75rem; font-weight: 800; }
.settings-list dd { min-width: 0; margin: 0; overflow-wrap: anywhere; color: #0f172a; font-size: 0.8125rem; line-height: 1.45; }
.settings-link, .settings-focus-button { display: inline-flex; width: fit-content; align-items: center; gap: 0.375rem; border: 1px solid #dbe4ee; border-radius: 0.625rem; background: #fff; padding: 0.5rem 0.75rem; color: #0f766e; font-size: 0.8125rem; font-weight: 800; text-decoration: none; cursor: pointer; }
.settings-link:hover, .settings-focus-button:hover { background: #f0fdfa; border-color: #99f6e4; }
.settings-chip-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.5rem; }
.settings-chip-grid span { display: grid; gap: 0.1875rem; border-radius: 0.625rem; background: #f8fafc; padding: 0.625rem 0.75rem; color: #0f172a; font-size: 0.8125rem; }
.settings-chip-grid b { color: #64748b; font-size: 0.6875rem; }
.dark .workbench-help-dialog, .dark .workbench-settings-dialog { background: #1c1917; border: 1px solid #292524; box-shadow: none; }
.dark .workbench-help-header { border-color: #292524; }
.dark .workbench-help-kicker { color: #5eead4; }
.dark .workbench-help-header h2, .dark .workbench-help-body strong, .dark .workbench-help-steps p, .dark .settings-section-title, .dark .settings-list dd, .dark .settings-chip-grid span { color: #fafaf9; }
.dark .workbench-help-close { color: #a8a29e; }
.dark .workbench-help-close:hover { background: #292524; color: #fafaf9; }
.dark .workbench-help-body span, .dark .workbench-help-steps ol, .dark .settings-section-compact p, .dark .settings-list dt, .dark .settings-chip-grid b { color: #a8a29e; }
.dark .workbench-help-grid article, .dark .workbench-help-steps, .dark .settings-section { border-color: #292524; background: #211d1a; }
.dark .settings-list div, .dark .settings-chip-grid span { background: #292524; }
.dark .settings-link, .dark .settings-focus-button { border-color: #44403c; background: #292524; color: #5eead4; }
.dark .settings-link:hover, .dark .settings-focus-button:hover { background: #1f2f2c; border-color: #0f766e; }
.dark .workbench-help-steps code { background: #312e81; color: #c4b5fd; }

@media (max-width: 640px) {
  .workbench-help-grid, .settings-chip-grid { grid-template-columns: 1fr; }
  .settings-list div { grid-template-columns: 1fr; gap: 0.25rem; }
}

.confirm-dialog { width: min(24rem, 92vw); background: white; border-radius: 1.5rem; padding: 1.5rem; }
.dark .confirm-dialog { background: #1c1917; border: 1px solid #292524; }
.btn-danger { background: #e11d48; color: white; border: none; padding: 0.5rem 1rem; border-radius: 0.5rem; cursor: pointer; font-weight: 500; font-size: 0.875rem; }
.btn-danger:hover { background: #be123c; }

/* ── Mobile ── */
@media (max-width: 640px) {
  .turn-image-grid { grid-template-columns: repeat(2, 1fr); }
  .turn-prompt-bubble { max-width: 92%; }
  .composer-inner { border-radius: 1.5rem; }
  .composer-toolbar-left { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .composer-actions { grid-column: 1 / -1; justify-content: flex-end; }
  .size-ratio-grid { grid-template-columns: repeat(2, 1fr); }
}

@media (min-width: 768px) {
  .image-workbench { height: calc(100dvh - 64px - 3rem); }
}

@media (min-width: 1024px) {
  .image-workbench { height: calc(100dvh - 64px - 4rem); }
}
</style>
