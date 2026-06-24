import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { createI18n } from "vue-i18n";
import SubscriptionPlanCard from "../SubscriptionPlanCard.vue";

const i18n = createI18n({
  legacy: false,
  locale: "en",
  fallbackWarn: false,
  missingWarn: false,
  messages: {
    en: {
      payment: {
        days: "days",
        models: "Models",
        planCard: {
          quota: "Quota",
          rate: "Rate",
          unlimited: "Unlimited",
        },
        currentBalance: "Current Balance",
        balancePayment: "Balance Payment",
        balanceAvailable: "Available",
        balanceShortfall: "Short {amount}",
        useBalanceSubscribe: "Use Balance",
        rechargeShortfallAction: "Recharge",
        externalPaymentSubscribe: "Pay Externally",
        subscribeNow: "Subscribe now",
      },
      common: {
        processing: "Processing...",
      },
    },
  },
});

const mountPlanCard = (groupPlatform: string) =>
  mount(SubscriptionPlanCard, {
    props: {
      plan: {
        id: 1,
        group_id: 10,
        group_platform: groupPlatform,
        name: "Pro",
        price: 10,
        amount: 1000,
        features: [],
        rate_multiplier: 1,
        validity_days: 30,
        validity_unit: "day",
        supported_model_scopes: ["claude", "gemini_text", "gemini_image"],
        is_active: true,
      },
    },
    global: { plugins: [i18n] },
  });

describe("SubscriptionPlanCard", () => {
  it("does not show Antigravity model scopes for OpenAI plans", () => {
    const text = mountPlanCard("openai").text();

    expect(text).not.toContain("Claude");
    expect(text).not.toContain("Gemini");
    expect(text).not.toContain("Imagen");
  });

  it("shows model scopes for Antigravity plans", () => {
    const text = mountPlanCard("antigravity").text();

    expect(text).toContain("Claude");
    expect(text).toContain("Gemini");
    expect(text).toContain("Imagen");
  });

  it("emits balance subscribe when balance covers the plan", async () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 1,
          group_id: 10,
          group_platform: "openai",
          name: "Pro",
          price: 10,
          amount: 1000,
          features: [],
          rate_multiplier: 1,
          validity_days: 30,
          validity_unit: "day",
          supported_model_scopes: [],
          is_active: true,
        },
        balance: 20,
        showBalanceAction: true,
      },
      global: { plugins: [i18n] },
    });

    expect(wrapper.text()).toContain("$20.00");
    expect(wrapper.text()).toContain("payment.useBalanceSubscribe");
    await wrapper.findAll("button")[0].trigger("click");

    expect(wrapper.emitted("balance-subscribe")?.[0]?.[0]).toMatchObject({ id: 1 });
    expect(wrapper.emitted("recharge")).toBeUndefined();
  });

  it("emits recharge when balance is short", async () => {
    const wrapper = mount(SubscriptionPlanCard, {
      props: {
        plan: {
          id: 1,
          group_id: 10,
          group_platform: "openai",
          name: "Pro",
          price: 10,
          amount: 1000,
          features: [],
          rate_multiplier: 1,
          validity_days: 30,
          validity_unit: "day",
          supported_model_scopes: [],
          is_active: true,
        },
        balance: 4,
        showBalanceAction: true,
      },
      global: { plugins: [i18n] },
    });

    expect(wrapper.text()).toContain("$4.00");
    expect(wrapper.text()).toContain("payment.rechargeShortfallAction");
    await wrapper.findAll("button")[0].trigger("click");

    expect(wrapper.emitted("recharge")?.[0]?.[0]).toMatchObject({ id: 1 });
    expect(wrapper.emitted("balance-subscribe")).toBeUndefined();
  });
});
