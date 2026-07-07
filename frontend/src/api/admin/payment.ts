/**
 * Admin Payment API endpoints
 * Handles payment management operations for administrators
 */

import { apiClient } from '../client'
import type {
  DashboardStats,
  PaymentOrder,
  PaymentChannel,
  SubscriptionPlan,
  ProviderInstance,
  ExternalFulfillmentSKU,
  ExternalOrderFulfillment,
  ExternalFulfillmentResult,
  UpsertExternalFulfillmentSKURequest,
  CreateExternalFulfillmentRequest
} from '@/types/payment'
import type { BasePaginationResponse } from '@/types'

/** Admin-facing payment config returned by GET /admin/payment/config */
export interface AdminPaymentConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  daily_limit: number
  order_timeout_minutes: number
  max_pending_orders: number
  enabled_payment_types: string[]
  balance_disabled: boolean
  balance_recharge_multiplier: number
  subscription_usd_to_cny_rate: number
  recharge_fee_rate: number
  load_balance_strategy: string
  product_name_prefix: string
  product_name_suffix: string
  help_image_url: string
  help_text: string
}

/** Fields accepted by PUT /admin/payment/config (all optional via pointer semantics) */
export interface UpdatePaymentConfigRequest {
  enabled?: boolean
  min_amount?: number
  max_amount?: number
  daily_limit?: number
  order_timeout_minutes?: number
  max_pending_orders?: number
  enabled_payment_types?: string[]
  balance_disabled?: boolean
  balance_recharge_multiplier?: number
  subscription_usd_to_cny_rate?: number
  recharge_fee_rate?: number
  load_balance_strategy?: string
  product_name_prefix?: string
  product_name_suffix?: string
  help_image_url?: string
  help_text?: string
}

export interface RefundResult {
  success: boolean
  warning?: string
  require_force?: boolean
  balance_deducted?: number
  subscription_days_deducted?: number
}

export const adminPaymentAPI = {
  // ==================== Config ====================

  /** Get payment configuration (admin view) */
  getConfig() {
    return apiClient.get<AdminPaymentConfig>('/admin/payment/config')
  },

  /** Update payment configuration */
  updateConfig(data: UpdatePaymentConfigRequest) {
    return apiClient.put('/admin/payment/config', data)
  },

  // ==================== Dashboard ====================

  /** Get payment dashboard statistics */
  getDashboard(days?: number) {
    return apiClient.get<DashboardStats>('/admin/payment/dashboard', {
      params: days ? { days } : undefined
    })
  },

  // ==================== Orders ====================

  /** Get all orders (paginated, with filters) */
  getOrders(params?: {
    page?: number
    page_size?: number
    status?: string
    payment_type?: string
    user_id?: number
    keyword?: string
    start_date?: string
    end_date?: string
    order_type?: string
  }) {
    return apiClient.get<BasePaginationResponse<PaymentOrder>>('/admin/payment/orders', { params })
  },

  /** Get a specific order by ID */
  getOrder(id: number) {
    return apiClient.get<PaymentOrder>(`/admin/payment/orders/${id}`)
  },

  /** Cancel an order (admin) */
  cancelOrder(id: number) {
    return apiClient.post(`/admin/payment/orders/${id}/cancel`)
  },

  /** Retry recharge for a failed order */
  retryRecharge(id: number) {
    return apiClient.post(`/admin/payment/orders/${id}/retry`)
  },

  /** Confirm a personal QR-code payment after manual receipt review */
  confirmPersonalQRCodeOrder(id: number, data: { amount: number; method: string; receipt_id: string; note?: string }) {
    return apiClient.post<PaymentOrder>(`/admin/payment/orders/${id}/confirm-personal-qrcode`, data)
  },

  /** Process a refund */
  refundOrder(id: number, data: { amount: number; reason: string; deduct_balance?: boolean; force?: boolean; manual_refund_reference?: string }) {
    return apiClient.post<RefundResult>(`/admin/payment/orders/${id}/refund`, data)
  },

  /** Query and finalize a pending refund */
  queryRefund(id: number) {
    return apiClient.post<RefundResult>(`/admin/payment/orders/${id}/refund/query`)
  },

  // ==================== Channels ====================

  /** Get all payment channels */
  getChannels() {
    return apiClient.get<PaymentChannel[]>('/admin/payment/channels')
  },

  /** Create a payment channel */
  createChannel(data: Partial<PaymentChannel>) {
    return apiClient.post<PaymentChannel>('/admin/payment/channels', data)
  },

  /** Update a payment channel */
  updateChannel(id: number, data: Partial<PaymentChannel>) {
    return apiClient.put<PaymentChannel>(`/admin/payment/channels/${id}`, data)
  },

  /** Delete a payment channel */
  deleteChannel(id: number) {
    return apiClient.delete(`/admin/payment/channels/${id}`)
  },

  // ==================== Subscription Plans ====================

  /** Get all subscription plans */
  getPlans() {
    return apiClient.get<SubscriptionPlan[]>('/admin/payment/plans')
  },

  /** Create a subscription plan */
  createPlan(data: Record<string, unknown>) {
    return apiClient.post<SubscriptionPlan>('/admin/payment/plans', data)
  },

  /** Update a subscription plan */
  updatePlan(id: number, data: Record<string, unknown>) {
    return apiClient.put<SubscriptionPlan>(`/admin/payment/plans/${id}`, data)
  },

  /** Delete a subscription plan */
  deletePlan(id: number) {
    return apiClient.delete(`/admin/payment/plans/${id}`)
  },

  // ==================== Provider Instances ====================

  /** Get all provider instances */
  getProviders() {
    return apiClient.get<ProviderInstance[]>('/admin/payment/providers')
  },

  /** Create a provider instance */
  createProvider(data: Partial<ProviderInstance>) {
    return apiClient.post<ProviderInstance>('/admin/payment/providers', data)
  },

  /** Update a provider instance */
  updateProvider(id: number, data: Partial<ProviderInstance>) {
    return apiClient.put<ProviderInstance>(`/admin/payment/providers/${id}`, data)
  },

  /** Delete a provider instance */
  deleteProvider(id: number) {
    return apiClient.delete(`/admin/payment/providers/${id}`)
  },

  // ==================== External Marketplace Fulfillment ====================

  /** List marketplace SKU mappings */
  listExternalFulfillmentSKUs(params?: {
    page?: number
    page_size?: number
    platform?: string
    enabled?: boolean
    keyword?: string
  }) {
    return apiClient.get<BasePaginationResponse<ExternalFulfillmentSKU>>('/admin/payment/external-fulfillment-skus', { params })
  },

  /** Create or update a marketplace SKU mapping */
  upsertExternalFulfillmentSKU(data: UpsertExternalFulfillmentSKURequest) {
    return apiClient.post<ExternalFulfillmentSKU>('/admin/payment/external-fulfillment-skus', data)
  },

  /** Delete a marketplace SKU mapping */
  deleteExternalFulfillmentSKU(id: number) {
    return apiClient.delete(`/admin/payment/external-fulfillment-skus/${id}`)
  },

  /** List marketplace order fulfillments */
  listExternalFulfillments(params?: {
    page?: number
    page_size?: number
    platform?: string
    status?: string
    sku_code?: string
    keyword?: string
    notify_status?: string
  }) {
    return apiClient.get<BasePaginationResponse<ExternalOrderFulfillment>>('/admin/payment/external-fulfillments', { params })
  },

  /** Create a fulfillment and redeem code */
  createExternalFulfillment(data: CreateExternalFulfillmentRequest) {
    return apiClient.post<ExternalFulfillmentResult>('/admin/payment/external-fulfillments', data)
  },

  /** Retry Feishu notification */
  retryExternalFulfillmentNotify(id: number) {
    return apiClient.post<ExternalOrderFulfillment>(`/admin/payment/external-fulfillments/${id}/retry-notify`)
  }
}

export default adminPaymentAPI
