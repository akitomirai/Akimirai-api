
<template>
  <div :class="[flat ? '' : (framed ? 'card overflow-hidden' : 'overflow-hidden'), dense ? 'usage-table-dense' : '']">
    <div
      v-if="showIpGeoToolbar"
      class="flex items-center justify-end gap-2 border-b border-gray-200 px-4 py-2 dark:border-dark-700"
    >
      <span v-if="pendingIpCount > 0" class="text-xs text-gray-500 dark:text-gray-400">
        {{ t('usage.ipGeo.pending', { count: pendingIpCount }) }}
      </span>
      <button
        type="button"
        class="inline-flex items-center gap-1 rounded px-2 py-1 text-xs font-medium text-primary-600 transition-colors hover:bg-primary-50 disabled:cursor-not-allowed disabled:opacity-50 dark:text-primary-400 dark:hover:bg-primary-900/30"
        :disabled="ipGeoBatchLoading || pendingIpCount === 0"
        @click="handleBatchFetchIpGeo"
      >
        {{ ipGeoBatchLoading ? t('usage.ipGeo.batchFetching') : t('usage.ipGeo.batchFetch') }}
      </button>
    </div>
    <div class="overflow-auto">
      <DataTable
        :columns="columns"
        :data="data"
        :loading="loading"
        :server-side-sort="serverSideSort"
        :default-sort-key="defaultSortKey"
        :default-sort-order="defaultSortOrder"
        :estimate-row-height="dense ? 52 : undefined"
        @sort="(key, order) => $emit('sort', key, order)"
      >
        <template #cell-user="{ row }">
          <div class="text-sm">
            <button
              v-if="row.user?.email"
              class="font-medium text-primary-600 underline decoration-dashed underline-offset-2 transition-colors hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
              @click="$emit('userClick', row.user_id, row.user?.email)"
              :title="t('admin.usage.clickToViewBalance')"
            >
              {{ row.user.email }}
            </button>
            <span v-else class="font-medium text-gray-900 dark:text-white">-</span>
            <span v-if="row.user?.deleted_at" class="ml-1 inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-rose-100 text-rose-600 ring-1 ring-inset ring-rose-200 dark:bg-rose-500/20 dark:text-rose-400 dark:ring-rose-500/30">
              {{ t('admin.usage.userDeletedBadge') }}
            </span>
            <span class="ml-1 text-gray-500 dark:text-gray-400">#{{ row.user_id }}</span>
          </div>
        </template>

        <template #cell-api_key="{ row }">
          <span class="text-sm text-gray-900 dark:text-white">{{ row.api_key?.name || '-' }}</span>
        </template>

        <template #cell-account="{ row }">
          <div class="min-w-[140px] space-y-1">
            <span class="text-sm text-gray-900 dark:text-white">{{ row.account?.name || '-' }}</span>
            <div v-if="row.route_kind" class="flex flex-wrap items-center gap-1 text-[11px]">
              <span class="rounded px-1.5 py-0.5 font-medium" :class="row.route_kind === 'proxy'
                ? 'bg-sky-100 text-sky-700 dark:bg-sky-500/20 dark:text-sky-300'
                : 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300'">
                {{ row.route_kind === 'proxy' ? (row.proxy_name_snapshot || t('admin.usage.diagnostics.proxyRoute')) : t('admin.usage.diagnostics.directRoute') }}
              </span>
              <span v-if="(row.retry_count ?? 0) > 0" class="rounded bg-amber-100 px-1.5 py-0.5 font-medium text-amber-700 dark:bg-amber-500/20 dark:text-amber-300">
                R{{ row.retry_count }}
              </span>
              <span v-if="(row.account_switch_count ?? 0) > 0" class="rounded bg-violet-100 px-1.5 py-0.5 font-medium text-violet-700 dark:bg-violet-500/20 dark:text-violet-300">
                S{{ row.account_switch_count }}
              </span>
            </div>
          </div>
        </template>

        <template #cell-model="{ row }">
          <div v-if="row.model_mapping_chain && row.model_mapping_chain.includes('→')" class="space-y-0.5 text-xs">
            <div v-for="(step, i) in row.model_mapping_chain.split('→')" :key="i"
                 class="flex max-w-[260px] flex-wrap items-center gap-1.5 whitespace-normal break-all"
                 :class="i === 0 ? 'font-medium text-gray-900 dark:text-white' : 'text-gray-500 dark:text-gray-400'"
                 :style="i > 0 ? `padding-left: ${i * 0.75}rem` : ''">
              <span v-if="i > 0" class="mr-0.5">→</span>
              <span>{{ step.trim() }}</span>
              <span
                v-if="i === 0 && getReasoningEffortBadgeLabel(row.reasoning_effort)"
                class="reasoning-effort-badge"
              >
                {{ getReasoningEffortBadgeLabel(row.reasoning_effort) }}
              </span>
            </div>
          </div>
          <div v-else-if="row.upstream_model && row.upstream_model !== row.model" class="space-y-0.5 text-xs">
            <div class="flex max-w-[260px] flex-wrap items-center gap-1.5 whitespace-normal break-all font-medium text-gray-900 dark:text-white">
              <span>{{ row.model }}</span>
              <span
                v-if="getReasoningEffortBadgeLabel(row.reasoning_effort)"
                class="reasoning-effort-badge"
              >
                {{ getReasoningEffortBadgeLabel(row.reasoning_effort) }}
              </span>
            </div>
            <div class="break-all text-gray-500 dark:text-gray-400">
              <span class="mr-0.5">→</span>{{ row.upstream_model }}
            </div>
          </div>
          <div v-else class="flex max-w-[260px] flex-wrap items-center gap-1.5 whitespace-normal break-all">
            <span class="font-medium text-gray-900 dark:text-white">{{ row.model }}</span>
            <span
              v-if="getReasoningEffortBadgeLabel(row.reasoning_effort)"
              class="reasoning-effort-badge"
            >
              {{ getReasoningEffortBadgeLabel(row.reasoning_effort) }}
            </span>
          </div>
        </template>

        <template #cell-reasoning_effort="{ row }">
          <span class="text-sm text-gray-900 dark:text-white">
            {{ formatReasoningEffort(row.reasoning_effort) }}
          </span>
        </template>

        <template #cell-endpoint="{ row }">
          <div v-if="showUpstreamEndpoint" class="max-w-[320px] space-y-1 text-xs">
            <div class="break-all text-gray-700 dark:text-gray-300">
              <span class="font-medium text-gray-500 dark:text-gray-400">{{ t('usage.inbound') }}:</span>
              <span class="ml-1">{{ row.inbound_endpoint?.trim() || '-' }}</span>
            </div>
            <div class="break-all text-gray-700 dark:text-gray-300">
              <span class="font-medium text-gray-500 dark:text-gray-400">{{ t('usage.upstream') }}:</span>
              <span class="ml-1">{{ row.upstream_endpoint?.trim() || '-' }}</span>
            </div>
          </div>
          <span
            v-else
            class="block max-w-[260px] truncate text-sm text-gray-700 dark:text-gray-300"
            :title="row.inbound_endpoint?.trim() || '-'"
          >
            {{ row.inbound_endpoint?.trim() || '-' }}
          </span>
        </template>

        <template #cell-group="{ row }">
          <span v-if="row.group" class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-200">
            {{ row.group.name }} · {{ formatMultiplier(row.group.rate_multiplier ?? 1) }}x
          </span>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #cell-stream="{ row }">
          <span class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium" :class="getRequestTypeBadgeClass(row)">
            {{ getRequestTypeLabel(row) }}
          </span>
        </template>

        <template #cell-billing_mode="{ row }">
          <span class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium" :class="getBillingModeBadgeClass(getDisplayBillingMode(row))">
            {{ getBillingModeLabel(getDisplayBillingMode(row), t) }}
          </span>
        </template>

        <template #cell-tokens="{ row }">
          <!-- 鍥剧墖鐢熸垚璇锋眰锛堜粎鎸夋璁¤垂鏃舵樉绀哄浘鐗囨牸寮忥級 -->
          <div v-if="isImageUsage(row)" class="flex items-center gap-1.5">
            <svg class="h-4 w-4 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <span class="font-medium text-gray-900 dark:text-white">{{ row.image_count }}{{ t('usage.imageUnit') }}</span>
            <span class="text-gray-400">({{ formatImageBillingSize(row, t) }})</span>
          </div>
          <!-- Token 璇锋眰 -->
          <div v-else-if="tokenBreakdown" class="flex items-center gap-1.5">
            <UsageTokenBreakdownCell :metrics="row" />
            <button
              type="button"
              class="group relative"
              data-testid="token-detail-trigger"
              :aria-label="t('usage.tokenDetails')"
              :aria-expanded="tokenTooltipVisible && tokenTooltipData === row && tokenTooltipMode === 'count'"
              @mouseenter="showTokenDetail($event, row)"
              @mouseleave="hideTokenTooltip"
              @click.stop.prevent
            >
              <div class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-gray-100 transition-colors group-hover:bg-blue-100 dark:bg-gray-700 dark:group-hover:bg-blue-900/50">
                <Icon name="infoCircle" size="xs" class="text-gray-400 group-hover:text-blue-500 dark:text-gray-500 dark:group-hover:text-blue-400" />
              </div>
            </button>
          </div>
          <div v-else class="flex items-center gap-1.5">
            <div class="space-y-1 text-sm">
              <span data-testid="usage-total-token-count" class="font-medium text-gray-900 dark:text-white">
                {{ formatTokenCount(totalTokenCount(row)) }}
              </span>
            </div>
            <!-- Token Detail Tooltip -->
            <button
              type="button"
              class="group relative"
              data-testid="token-detail-trigger"
              :aria-label="t('usage.tokenDetails')"
              :aria-expanded="tokenTooltipVisible && tokenTooltipData === row && tokenTooltipMode === 'count'"
              @mouseenter="showTokenDetail($event, row)"
              @mouseleave="hideTokenTooltip"
              @click.stop.prevent
            >
              <div class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-gray-100 transition-colors group-hover:bg-blue-100 dark:bg-gray-700 dark:group-hover:bg-blue-900/50">
                <Icon name="infoCircle" size="xs" class="text-gray-400 group-hover:text-blue-500 dark:text-gray-500 dark:group-hover:text-blue-400" />
              </div>
            </button>
          </div>
        </template>

        <template #cell-cost="{ row }">
          <div class="text-sm">
            <div class="flex items-center gap-1.5">
              <span class="font-medium text-green-600 dark:text-green-400">${{ row.actual_cost?.toFixed(6) || '0.000000' }}</span>
              <span
                v-if="row.long_context_billing_applied"
                data-testid="long-context-billing-marker"
                class="inline-flex items-center rounded px-1 py-px text-[10px] font-semibold leading-tight bg-amber-100 text-amber-700 ring-1 ring-inset ring-amber-200 dark:bg-amber-500/20 dark:text-amber-300 dark:ring-amber-500/30"
              >x2</span>
              <button
                type="button"
                class="group relative"
                data-testid="cost-detail-trigger"
                :aria-label="t('usage.costDetails')"
                :aria-expanded="isImageUsage(row)
                  ? tooltipVisible && tooltipData === row
                  : tokenTooltipVisible && tokenTooltipData === row && tokenTooltipMode === 'cost'"
                @mouseenter="showCostDetail($event, row)"
                @mouseleave="hideCostDetail(row)"
                @click.stop.prevent
              >
                <div class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-gray-100 transition-colors group-hover:bg-blue-100 dark:bg-gray-700 dark:group-hover:bg-blue-900/50">
                  <Icon name="infoCircle" size="xs" class="text-gray-400 group-hover:text-blue-500 dark:text-gray-500 dark:group-hover:text-blue-400" />
                </div>
              </button>
            </div>
            <div v-if="showAccountBilling && row.account_rate_multiplier != null" class="mt-0.5 text-[11px] text-orange-500 dark:text-orange-400">
              A ${{ accountBilled(row).toFixed(6) }}
            </div>
          </div>
        </template>


        <template #cell-cache_hit_rate="{ row }">
          <UsageCacheHitCell :metrics="row" />
        </template>

        <template #cell-first_token="{ row }">
          <span v-if="row.first_token_ms != null" class="text-sm text-gray-600 dark:text-gray-400">{{ formatDuration(row.first_token_ms) }}</span>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #cell-duration="{ row }">
          <span class="text-sm text-gray-600 dark:text-gray-400">{{ formatDuration(row.duration_ms) }}</span>
        </template>

        <!-- Combined first-token/duration health column for scan-friendly latency review. -->
        <template #cell-latency="{ row }">
          <UsageLatencyCell
            :first-token-ms="displayFirstTokenMs(row)"
            :duration-ms="displayTotalMs(row)"
          />
        </template>

        <template #cell-created_at="{ value }">
          <span class="text-sm text-gray-600 dark:text-gray-400">{{ formatDateTime(value) }}</span>
        </template>

        <template #cell-user_agent="{ row }">
          <span v-if="row.user_agent" class="text-sm text-gray-600 dark:text-gray-400 block max-w-[320px] truncate" :title="row.user_agent">{{ formatUserAgent(row.user_agent) }}</span>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #cell-ip_address="{ row }">
          <div v-if="row.ip_address">
            <span class="text-sm font-mono text-gray-600 dark:text-gray-400">{{ row.ip_address }}</span>
            <IpGeoCell :ip="row.ip_address" />
          </div>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #cell-diagnostics="{ row }">
          <button
            type="button"
            class="flex h-8 w-8 items-center justify-center rounded-md text-gray-500 transition-colors hover:bg-primary-50 hover:text-primary-600 focus:outline-none focus:ring-2 focus:ring-primary-500 dark:hover:bg-primary-500/10 dark:hover:text-primary-300"
            data-testid="diagnostics-trigger"
            :title="t('admin.usage.diagnostics.open')"
            @click="$emit('diagnosticsClick', row.id)"
          >
            <Icon name="search" size="sm" />
          </button>
        </template>

      </DataTable>
    </div>
  </div>

  <!-- Token Tooltip Portal -->
  <Teleport to="body">
    <div
      v-if="tokenTooltipVisible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{
        left: tokenTooltipPosition.x + 'px',
        top: tokenTooltipPosition.y + 'px'
      }"
    >
      <div
        class="max-w-[calc(100vw-32px)] overflow-x-auto rounded-lg border border-gray-700 bg-gray-900 px-3 py-2.5 text-xs text-white shadow-xl dark:border-gray-600 dark:bg-gray-800"
        :class="tokenTooltipMode === 'cost' ? 'w-[460px]' : 'w-[240px]'"
      >
        <div v-if="tokenTooltipMode === 'cost'" class="space-y-1.5">
          <div class="usage-cost-grid border-b border-gray-700 pb-1.5 text-gray-300">
            <span class="col-span-2 font-semibold">{{ t('usage.tokenDetails') }}</span>
            <span>{{ t('usage.tokenUnitPrice') }}</span>
            <span>{{ t('usage.tokenCost') }}</span>
          </div>
          <div data-testid="cost-input-token-row" class="usage-cost-grid">
            <span class="inline-flex items-center gap-1.5 text-gray-400">
              <Icon data-testid="cost-input-token-icon" name="arrowDown" size="xs" class="shrink-0 text-emerald-400" />
              {{ t('admin.usage.inputTokens') }}
            </span>
            <span class="font-medium tabular-nums text-emerald-300">{{ formatTokenCount(textInputTokens(tokenTooltipData), { display: 'exact' }) }}</span>
            <span class="font-medium text-sky-300">{{ formatTokenUnitPrice(tokenTooltipData?.input_cost, textInputTokens(tokenTooltipData)) }}</span>
            <span class="font-medium tabular-nums text-white">${{ tokenTooltipData?.input_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <div v-if="tokenTooltipData && hasImageInputTokens(tokenTooltipData)" data-testid="cost-image-input-token-row" class="usage-cost-grid">
            <span class="col-span-2 text-gray-400">{{ t('usage.imageInputTokens') }}</span>
            <span class="font-medium text-fuchsia-300">{{ formatTokenUnitPrice(tokenTooltipData.image_input_cost, tokenTooltipData.image_input_tokens) }}</span>
            <span class="font-medium tabular-nums text-fuchsia-300">${{ tokenTooltipData.image_input_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <div data-testid="cost-output-token-row" class="usage-cost-grid">
            <span class="inline-flex items-center gap-1.5 text-gray-400">
              <Icon data-testid="cost-output-token-icon" name="arrowUp" size="xs" class="shrink-0 text-violet-400" />
              {{ t('admin.usage.outputTokens') }}
            </span>
            <span class="font-medium tabular-nums text-violet-300">{{ formatTokenCount(textOutputTokens(tokenTooltipData), { display: 'exact' }) }}</span>
            <span class="font-medium text-violet-300">{{ formatTokenUnitPrice(tokenTooltipData?.output_cost, textOutputTokens(tokenTooltipData)) }}</span>
            <span class="font-medium tabular-nums text-white">${{ tokenTooltipData?.output_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <div v-if="tokenTooltipData && hasImageOutputTokens(tokenTooltipData)" data-testid="cost-image-output-token-row" class="usage-cost-grid">
            <span class="col-span-2 text-gray-400">{{ t('usage.imageOutputTokens') }}</span>
            <span class="font-medium text-pink-300">{{ formatTokenUnitPrice(tokenTooltipData.image_output_cost, tokenTooltipData.image_output_tokens) }}</span>
            <span class="font-medium tabular-nums text-pink-300">${{ tokenTooltipData.image_output_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <div data-testid="cost-cache-token-row" class="usage-cost-grid">
            <span class="inline-flex items-center gap-1.5 text-gray-400">
              <Icon data-testid="cost-cache-token-icon" name="database" size="xs" class="shrink-0 text-sky-400" />
              {{ t('admin.usage.cacheReadTokens') }}
            </span>
            <span class="font-medium tabular-nums text-sky-300">{{ formatTokenCount(tokenTooltipData?.cache_read_tokens ?? 0, { display: 'exact' }) }}</span>
            <span class="font-medium text-sky-300">{{ formatTokenUnitPrice(tokenTooltipData?.cache_read_cost, tokenTooltipData?.cache_read_tokens) }}</span>
            <span class="font-medium tabular-nums text-white">${{ tokenTooltipData?.cache_read_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <div class="usage-cost-grid border-t border-gray-700 pt-1.5">
            <span class="text-gray-400">{{ t('usage.totalTokens') }}</span>
            <span class="font-semibold text-blue-400">{{ formatTokenCount(totalTokenCount(tokenTooltipData), { display: 'exact' }) }}</span>
            <span></span>
            <span class="font-semibold text-white">${{ tokenTooltipData?.total_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <div v-if="tokenTooltipData && cacheHitRatio(tokenTooltipData) !== null" class="usage-cost-grid">
            <span class="col-span-3 text-gray-400">{{ t('usage.cacheHitRate') }}</span>
            <span class="font-semibold text-sky-400">{{ cacheHitRatio(tokenTooltipData) }}%</span>
          </div>
          <div class="usage-cost-grid border-t border-gray-700 pt-1.5">
            <span class="col-span-3 text-gray-400">{{ t('usage.rate') }}</span>
            <span class="font-semibold text-blue-400">{{ formatMultiplier(tokenTooltipData?.rate_multiplier || 1) }}x</span>
          </div>
          <div class="usage-cost-grid">
            <span class="col-span-3 text-gray-400">{{ t('usage.userBilled') }}</span>
            <span class="font-semibold text-green-400">${{ tokenTooltipData?.actual_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <div class="usage-cost-grid">
            <span class="col-span-3 text-gray-400">{{ t('usage.serviceTier') }}</span>
            <span class="font-semibold text-cyan-300">{{ getUsageServiceTierLabel(tokenTooltipData?.service_tier, t) }}</span>
          </div>
          <template v-if="showAccountBilling">
            <div class="usage-cost-grid border-t border-gray-700 pt-1.5">
              <span class="col-span-3 text-gray-400">{{ t('usage.accountMultiplier') }}</span>
              <span class="font-semibold text-blue-400">{{ formatMultiplier(tokenTooltipData?.account_rate_multiplier ?? 1) }}x</span>
            </div>
            <div class="usage-cost-grid">
              <span class="col-span-3 text-gray-400">{{ t('usage.accountBilled') }}</span>
              <span class="font-semibold text-green-400">${{ accountBilled({
                total_cost: tokenTooltipData?.total_cost,
                account_stats_cost: tokenTooltipData?.account_stats_cost,
                account_rate_multiplier: tokenTooltipData?.account_rate_multiplier,
              }).toFixed(6) }}</span>
            </div>
          </template>
        </div>
        <div v-else class="space-y-1.5">
          <div class="text-xs font-semibold text-gray-300">{{ t('usage.tokenDetails') }}</div>
          <div class="flex items-center justify-between gap-4 border-t border-gray-700 pt-1.5">
            <span class="text-gray-400">{{ t('admin.usage.inputTokens') }}</span>
            <span class="font-medium text-white">{{ formatTokenCount(textInputTokens(tokenTooltipData), { display: 'exact' }) }}</span>
          </div>
          <div v-if="tokenTooltipData && hasImageInputTokens(tokenTooltipData)" class="flex items-center justify-between gap-4">
            <span class="text-gray-400">{{ t('usage.imageInputTokens') }}</span>
            <span class="font-medium text-fuchsia-300">{{ formatTokenCount(tokenTooltipData.image_input_tokens, { display: 'exact' }) }}</span>
          </div>
          <div class="flex items-center justify-between gap-4">
            <span class="text-gray-400">{{ t('admin.usage.outputTokens') }}</span>
            <span class="font-medium text-white">{{ formatTokenCount(textOutputTokens(tokenTooltipData), { display: 'exact' }) }}</span>
          </div>
          <div v-if="tokenTooltipData && hasImageOutputTokens(tokenTooltipData)" class="flex items-center justify-between gap-4">
            <span class="text-gray-400">{{ t('usage.imageOutputTokens') }}</span>
            <span class="font-medium text-pink-300">{{ formatTokenCount(tokenTooltipData.image_output_tokens, { display: 'exact' }) }}</span>
          </div>
          <div v-if="tokenTooltipData && tokenTooltipData.cache_creation_tokens > 0" class="flex items-center justify-between gap-4">
            <span class="text-gray-400">{{ t('admin.usage.cacheCreationTokens') }}</span>
            <span class="font-medium text-white">{{ formatTokenCount(tokenTooltipData.cache_creation_tokens, { display: 'exact' }) }}</span>
          </div>
          <div class="flex items-center justify-between gap-4">
            <span class="text-gray-400">{{ t('admin.usage.cacheReadTokens') }}</span>
            <span class="font-medium text-white">{{ formatTokenCount(tokenTooltipData?.cache_read_tokens ?? 0, { display: 'exact' }) }}</span>
          </div>
          <div class="flex items-center justify-between gap-4 border-t border-gray-700 pt-1.5">
            <span class="text-gray-400">{{ t('usage.totalTokens') }}</span>
            <span class="font-semibold text-blue-400">{{ formatTokenCount(totalTokenCount(tokenTooltipData), { display: 'exact' }) }}</span>
          </div>
          <div v-if="tokenTooltipData && cacheHitRatio(tokenTooltipData) !== null" class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.cacheHitRate') }}</span>
            <span class="font-semibold text-sky-400">{{ cacheHitRatio(tokenTooltipData) }}%</span>
          </div>
        </div>
        <div
          class="absolute top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-t-[6px] border-b-transparent border-t-transparent"
          :class="tokenTooltipSide === 'right'
            ? 'right-full border-r-[6px] border-r-gray-900 dark:border-r-gray-800'
            : 'left-full border-l-[6px] border-l-gray-900 dark:border-l-gray-800'"
        ></div>
      </div>
    </div>
  </Teleport>

  <!-- Cost Tooltip Portal -->
  <Teleport to="body">
    <div
      v-if="tooltipVisible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{
        left: tooltipPosition.x + 'px',
        top: tooltipPosition.y + 'px'
      }"
    >
      <div class="whitespace-nowrap rounded-lg border border-gray-700 bg-gray-900 px-3 py-2.5 text-xs text-white shadow-xl dark:border-gray-600 dark:bg-gray-800">
        <div class="space-y-1.5">
          <!-- Cost Breakdown -->
          <div class="mb-2 border-b border-gray-700 pb-1.5">
            <div class="text-xs font-semibold text-gray-300 mb-1">{{ t('usage.costDetails') }}</div>
            <div v-if="tooltipData && tooltipData.input_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.inputCost') }}</span>
              <span class="font-medium text-white">${{ tooltipData.input_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && hasImageInputCost(tooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.imageInputCost') }}</span>
              <span class="font-medium text-fuchsia-300">${{ tooltipData.image_input_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.output_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.outputCost') }}</span>
              <span class="font-medium text-white">${{ tooltipData.output_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && hasImageOutputCost(tooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.imageOutputCost') }}</span>
              <span class="font-medium text-pink-300">${{ tooltipData.image_output_cost.toFixed(6) }}</span>
            </div>
            <!-- Token billing: show unit prices per 1M tokens -->
            <template v-if="tooltipData && !isImageUsage(tooltipData) && (!tooltipData.billing_mode || tooltipData.billing_mode === BILLING_MODE_TOKEN)">
              <div v-if="tooltipData && textInputTokens(tooltipData) > 0" class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.inputTokenPrice') }}</span>
                <span class="font-medium text-sky-300">{{ formatTokenPricePerMillion(tooltipData.input_cost, textInputTokens(tooltipData)) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && hasImageInputTokens(tooltipData)" class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageInputTokenPrice') }}</span>
                <span class="font-medium text-fuchsia-300">{{ formatTokenPricePerMillion(tooltipData.image_input_cost ?? 0, tooltipData.image_input_tokens) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && tooltipData.output_cost > 0 && textOutputTokens(tooltipData) > 0" class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.outputTokenPrice') }}</span>
                <span class="font-medium text-violet-300">{{ formatTokenPricePerMillion(tooltipData.output_cost, textOutputTokens(tooltipData)) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && tooltipData.cache_read_tokens > 0" class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.cacheReadTokenPrice') }}</span>
                <span class="font-medium text-sky-300">{{ formatTokenPricePerMillion(tooltipData.cache_read_cost, tooltipData.cache_read_tokens) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && hasImageOutputTokens(tooltipData)" class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageOutputTokenPrice') }}</span>
                <span class="font-medium text-pink-300">{{ formatTokenPricePerMillion(tooltipData.image_output_cost ?? 0, tooltipData.image_output_tokens) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
            </template>
            <template v-else-if="tooltipData && isImageUsage(tooltipData)">
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageCount') }}</span>
                <span class="font-medium text-white">{{ tooltipData.image_count }}{{ t('usage.imageUnit') }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageBillingSize') }}</span>
                <span class="font-medium text-white">{{ formatImageBillingSize(tooltipData, t) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageSizeSource') }}</span>
                <span class="font-medium text-white">{{ formatImageSizeSource(tooltipData, t) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageInputSize') }}</span>
                <span class="font-medium text-white">{{ formatImageInputSize(tooltipData, t) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageOutputSize') }}</span>
                <span class="font-medium text-white">{{ formatImageOutputSize(tooltipData, t) }}</span>
              </div>
              <div v-if="formatImageSizeBreakdown(tooltipData)" class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageSizeBreakdown') }}</span>
                <span class="font-medium text-white">{{ formatImageSizeBreakdown(tooltipData) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageUnitPrice') }}</span>
                <span class="font-medium text-sky-300">${{ imageUnitPrice(tooltipData).toFixed(6) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-gray-400">{{ t('usage.imageTotalPrice') }}</span>
                <span class="font-medium text-white">${{ tooltipData.total_cost?.toFixed(6) || '0.000000' }}</span>
              </div>
            </template>
            <div v-else class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.unitPrice') }}</span>
              <span class="font-medium text-sky-300">${{ tooltipData?.total_cost?.toFixed(6) || '0.000000' }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.cache_creation_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheCreationCost') }}</span>
              <span class="font-medium text-white">${{ tooltipData.cache_creation_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.cache_read_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.cacheReadCost') }}</span>
              <span class="font-medium text-white">${{ tooltipData.cache_read_cost.toFixed(6) }}</span>
            </div>
          </div>
          <!-- Rate and Summary -->
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.serviceTier') }}</span>
            <span class="font-semibold text-cyan-300">{{ getUsageServiceTierLabel(tooltipData?.service_tier, t) }}</span>
          </div>
          <!-- Latency Breakdown -->
          <div v-if="hasLatencyBreakdown(tooltipData)" class="mb-2 border-b border-gray-700 pb-1.5">
            <div class="text-xs font-semibold text-gray-300 mb-1">{{ t('usage.latencyBreakdown') }}</div>
            <div v-if="tooltipData?.client_transport" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.clientTransport') }}</span>
              <span class="font-medium"
                :class="tooltipData.client_transport === 'ws' ? 'text-violet-300' : 'text-emerald-300'"
              >{{ tooltipData.client_transport.toUpperCase() }}</span>
            </div>
            <div v-if="tooltipData?.auth_latency_ms != null" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.authLatency') }}</span>
              <span class="font-medium text-white">{{ formatDuration(tooltipData.auth_latency_ms) }}</span>
            </div>
            <div v-if="tooltipData?.routing_latency_ms != null" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.routingLatency') }}</span>
              <span class="font-medium text-white">{{ formatDuration(tooltipData.routing_latency_ms) }}</span>
            </div>
            <div v-if="tooltipData?.upstream_latency_ms != null" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.upstreamLatency') }}</span>
              <span class="font-medium text-white">{{ formatDuration(tooltipData.upstream_latency_ms) }}</span>
            </div>
            <div v-if="tooltipData?.response_latency_ms != null" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.responseLatency') }}</span>
              <span class="font-medium text-white">{{ formatDuration(tooltipData.response_latency_ms) }}</span>
            </div>
            <div v-if="tooltipData?.request_body_read_ms != null" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.diagnostics.bodyRead') }}</span>
              <span class="font-medium text-white">{{ formatDuration(tooltipData.request_body_read_ms) }}</span>
            </div>
            <div v-if="tooltipData?.upstream_request_written_ms != null" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.diagnostics.requestWritten') }}</span>
              <span class="font-medium text-white">{{ formatDuration(tooltipData.upstream_request_written_ms) }}</span>
            </div>
            <div v-if="tooltipData?.upstream_first_byte_ms != null" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.diagnostics.firstByte') }}</span>
              <span class="font-medium text-white">{{ formatDuration(tooltipData.upstream_first_byte_ms) }}</span>
            </div>
            <div v-if="tooltipData?.request_first_token_ms != null" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.diagnostics.firstToken') }}</span>
              <span class="font-medium text-amber-300">{{ formatDuration(tooltipData.request_first_token_ms) }}</span>
            </div>
            <div v-if="tooltipData?.request_total_ms != null" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('admin.usage.diagnostics.totalDuration') }}</span>
              <span class="font-medium text-white">{{ formatDuration(tooltipData.request_total_ms) }}</span>
            </div>
            <div v-if="tooltipData?.duration_ms != null" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.duration') }}</span>
              <span class="font-medium text-white">{{ formatDuration(tooltipData.duration_ms) }}</span>
            </div>
            <div v-if="tooltipData?.first_token_ms != null" class="flex items-center justify-between gap-4">
              <span class="text-gray-400">{{ t('usage.firstToken') }}</span>
              <span class="font-medium text-amber-300">{{ formatDuration(tooltipData.first_token_ms) }}</span>
            </div>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.rate') }}</span>
            <span class="font-semibold text-blue-400">{{ formatMultiplier(tooltipData?.rate_multiplier || 1) }}x</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.original') }}</span>
            <span class="font-medium text-white">${{ tooltipData?.total_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-gray-400">{{ t('usage.userBilled') }}</span>
            <span class="font-semibold text-green-400">${{ tooltipData?.actual_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <!-- Account billing (separated from user billing) -->
          <template v-if="showAccountBilling">
            <div class="flex items-center justify-between gap-6 border-t border-gray-700 pt-1.5">
              <span class="text-gray-400">{{ t('usage.accountMultiplier') }}</span>
              <span class="font-semibold text-blue-400">{{ formatMultiplier(tooltipData?.account_rate_multiplier ?? 1) }}x</span>
            </div>
            <div class="flex items-center justify-between gap-6">
              <span class="text-gray-400">{{ t('usage.accountBilled') }}</span>
              <span class="font-semibold text-green-400">
                ${{ accountBilled({
                  total_cost: tooltipData?.total_cost,
                  account_stats_cost: tooltipData?.account_stats_cost,
                  account_rate_multiplier: tooltipData?.account_rate_multiplier,
                }).toFixed(6) }}
              </span>
            </div>
          </template>
        </div>
        <div class="absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-r-gray-900 border-t-transparent dark:border-r-gray-800"></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatDateTime, formatReasoningEffort, formatTokenCount } from '@/utils/format'
import { formatMultiplier } from '@/utils/formatters'
import { formatTokenPricePerMillion } from '@/utils/usagePricing'
import { getUsageServiceTierLabel } from '@/utils/usageServiceTier'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import { cacheHitRatio } from '@/utils/usageMetrics'
import {
  BILLING_MODE_TOKEN,
  getBillingModeLabel,
  getBillingModeBadgeClass,
  isImageUsage,
  getDisplayBillingMode,
  imageUnitPrice,
} from '@/utils/billingMode'
import {
  formatImageBillingSize,
  formatImageInputSize,
  formatImageOutputSize,
  formatImageSizeBreakdown,
  formatImageSizeSource,
  hasImageOutputTokens,
  textOutputTokens,
  hasImageOutputCost,
  hasImageInputTokens,
  textInputTokens,
  hasImageInputCost,
} from '@/utils/imageUsage'

/** Compute the account-billed cost for display: (account_stats_cost ?? total_cost) * rate_multiplier */
function accountBilled(row: { total_cost?: number | null; account_stats_cost?: number | null; account_rate_multiplier?: number | null }): number {
  const base = row.account_stats_cost != null ? row.account_stats_cost : (row.total_cost ?? 0)
  const result = base * (row.account_rate_multiplier ?? 1)
  return Number.isNaN(result) ? 0 : result
}


import DataTable from '@/components/common/DataTable.vue'
import IpGeoCell from '@/components/common/IpGeoCell.vue'
import Icon from '@/components/icons/Icon.vue'
import UsageCacheHitCell from '@/components/usage/UsageCacheHitCell.vue'
import UsageLatencyCell from '@/components/usage/UsageLatencyCell.vue'
import UsageTokenBreakdownCell from '@/components/usage/UsageTokenBreakdownCell.vue'
import { fetchBatch, getEntry } from '@/utils/ipGeoLookup'
import type { AdminUsageLog } from '@/types'
import type { Column } from '@/components/common/types'

interface Props {
  data: AdminUsageLog[]
  loading?: boolean
  columns: Column[]
  serverSideSort?: boolean
  defaultSortKey?: string
  defaultSortOrder?: 'asc' | 'desc'
  showAccountBilling?: boolean
  showUpstreamEndpoint?: boolean
  tokenBreakdown?: boolean

  dense?: boolean
  framed?: boolean
  /** Embedded usage inside a shared card: remove this component's card shell. */
  flat?: boolean

}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  serverSideSort: false,
  defaultSortKey: '',
  defaultSortOrder: 'asc',
  showAccountBilling: true,
  showUpstreamEndpoint: true,
  tokenBreakdown: false,

  dense: false,
  framed: true,
  flat: false

})
const emit = defineEmits<{
  userClick: [userID: number, email?: string]
  sort: [key: string, order: 'asc' | 'desc']
  ipGeoBatchFailed: []
  diagnosticsClick: [usageID: number]
}>()
const { t } = useI18n()
const showAccountBilling = props.showAccountBilling
const showUpstreamEndpoint = props.showUpstreamEndpoint
const tokenBreakdown = props.tokenBreakdown
const ipGeoBatchLoading = ref(false)

const showIpGeoToolbar = computed(() => props.columns.some((col) => col.key === 'ip_address'))

const currentPageIps = computed(() =>
  Array.from(new Set(props.data.map((row) => row.ip_address).filter((ip): ip is string => Boolean(ip))))
)

const pendingIpCount = computed(() => {
  if (!showIpGeoToolbar.value) return 0
  return currentPageIps.value.filter((ip) => {
    const status = getEntry(ip).status
    return status === 'idle' || status === 'error'
  }).length
})

const handleBatchFetchIpGeo = async () => {
  ipGeoBatchLoading.value = true
  try {
    const ok = await fetchBatch(currentPageIps.value)
    if (!ok) emit('ipGeoBatchFailed')
  } finally {
    ipGeoBatchLoading.value = false
  }
}

// Tooltip state - cost
const tooltipVisible = ref(false)
const tooltipPosition = ref({ x: 0, y: 0 })
const tooltipData = ref<AdminUsageLog | null>(null)

// Tooltip state - token
const tokenTooltipVisible = ref(false)
const tokenTooltipPosition = ref({ x: 0, y: 0 })
const tokenTooltipData = ref<AdminUsageLog | null>(null)
const tokenTooltipMode = ref<'cost' | 'count'>('count')
const tokenTooltipSide = ref<'left' | 'right'>('right')

const getRequestTypeLabel = (row: AdminUsageLog): string => {
  const requestType = resolveUsageRequestType(row)
  if (requestType === 'cyber') return t('usage.cyber')
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
}

const getRequestTypeBadgeClass = (row: AdminUsageLog): string => {
  const requestType = resolveUsageRequestType(row)
  if (requestType === 'cyber') return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  if (requestType === 'ws_v2') return 'bg-violet-100 text-violet-800 dark:bg-violet-900 dark:text-violet-200'
  if (requestType === 'stream') return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
  if (requestType === 'sync') return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
  return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200'
}

const totalTokenCount = (row: AdminUsageLog | null | undefined): number => {
  if (!row) return 0
  return (row.input_tokens || 0)
    + (row.output_tokens || 0)
    + (row.cache_creation_tokens || 0)
    + (row.cache_read_tokens || 0)
}



const formatUserAgent = (ua: string): string => {
  return ua
}

// 超过 1 分钟简化为 "Xm Ys"，免去人工换算（超过 1 小时再进位为 "Xh Ym"）
const formatDuration = (ms: number | null | undefined): string => {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(2)}s`
  const totalSec = Math.round(ms / 1000)
  if (totalSec < 3600) return `${Math.floor(totalSec / 60)}m ${totalSec % 60}s`
  return `${Math.floor(totalSec / 3600)}h ${Math.floor((totalSec % 3600) / 60)}m`
}

const formatTokenUnitPrice = (cost: number | null | undefined, tokens: number | null | undefined): string => {
  const price = formatTokenPricePerMillion(cost, tokens)
  return price === '-' ? price : `${price} ${t('usage.perMillionTokens')}`
}

const displayFirstTokenMs = (row: AdminUsageLog): number | null => row.request_first_token_ms ?? row.first_token_ms ?? null
const displayTotalMs = (row: AdminUsageLog): number | null => row.request_total_ms ?? row.duration_ms ?? null

const getReasoningEffortBadgeLabel = (effort: string | null | undefined): string => {
  const label = formatReasoningEffort(effort)
  return label === '-' ? '' : label.toLowerCase()
}

// Cost tooltip functions. Image billing keeps its existing detail layout.
const showTooltip = (event: MouseEvent, row: AdminUsageLog) => {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  tooltipData.value = row
  tooltipPosition.value.x = rect.right + 8
  tooltipPosition.value.y = rect.top + rect.height / 2
  tooltipVisible.value = true
}

const hideTooltip = () => {
  tooltipVisible.value = false
  tooltipData.value = null
}

// Token tooltip functions. Token billing uses the cost icon for the price/cost view;
// the Token icon uses the same anchored popover in count-only mode.
const showTokenTooltip = (event: MouseEvent, row: AdminUsageLog, mode: 'cost' | 'count') => {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  tokenTooltipData.value = row
  tokenTooltipMode.value = mode
  const tooltipWidth = mode === 'cost' ? 460 : 240
  const viewportWidth = window.innerWidth || document.documentElement.clientWidth
  const preferredRight = rect.right + 8
  const canPlaceRight = preferredRight + tooltipWidth <= viewportWidth - 16
  tokenTooltipSide.value = canPlaceRight ? 'right' : 'left'
  tokenTooltipPosition.value.x = canPlaceRight ? preferredRight : Math.max(16, rect.left - tooltipWidth - 8)
  tokenTooltipPosition.value.y = rect.top + rect.height / 2
  tokenTooltipVisible.value = true
}

const hideTokenTooltip = () => {
  tokenTooltipVisible.value = false
  tokenTooltipData.value = null
}

const showTokenDetail = (event: MouseEvent, row: AdminUsageLog) => {
  hideTooltip()
  showTokenTooltip(event, row, 'count')
}

const showCostDetail = (event: MouseEvent, row: AdminUsageLog) => {
  if (isImageUsage(row)) {
    hideTokenTooltip()
    showTooltip(event, row)
    return
  }
  hideTooltip()
  showTokenTooltip(event, row, 'cost')
}

const hideCostDetail = (row: AdminUsageLog) => {
  if (isImageUsage(row)) {
    hideTooltip()
    return
  }
  hideTokenTooltip()
}

/** 鏄惁鍖呭惈寤惰繜鍒嗚В淇℃伅 */
const hasLatencyBreakdown = (row: AdminUsageLog | null | undefined): boolean => {
  if (!row) return false
  return (
    row.client_transport != null ||
    row.auth_latency_ms != null ||
    row.routing_latency_ms != null ||
    row.upstream_latency_ms != null ||
    row.response_latency_ms != null ||
    row.request_body_read_ms != null ||
    row.upstream_request_written_ms != null ||
    row.upstream_first_byte_ms != null ||
    row.request_first_token_ms != null ||
    row.request_total_ms != null
  )
}

/** 璁＄畻鍗曡缂撳瓨鍛戒腑鐜?= cache_read / (input + cache_read + cache_write) 脳 100 */
</script>

<style scoped>
.usage-cost-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1.15fr) auto;
  align-items: center;
  column-gap: 1rem;
}

.usage-table-dense :deep(thead th) {
  padding-top: 0.625rem;
  padding-bottom: 0.625rem;
}

.usage-table-dense :deep(tbody td) {
  padding-top: 0.75rem;
  padding-bottom: 0.75rem;
  vertical-align: middle;
}

.reasoning-effort-badge {
  display: inline-flex;
  height: 1.25rem;
  align-items: center;
  border-radius: 9999px;
  padding: 0 0.375rem;
  font-size: 11px;
  font-weight: 600;
  line-height: 1;
  color: rgb(249 115 22);
  background-color: rgb(255 247 237);
  box-shadow: inset 0 0 0 1px rgb(255 237 213);
}

.dark .reasoning-effort-badge {
  color: rgb(253 186 116);
  background-color: rgb(249 115 22 / 0.1);
  box-shadow: inset 0 0 0 1px rgb(249 115 22 / 0.2);
}
</style>
