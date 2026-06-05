import axios from "axios";
import { For, createSignal, onMount } from "solid-js";
import { useI18n } from "../i18n";
import { formatLocalDateTime, formatLocalDate } from "../utils/format";

const MenuPricing = () => {
  const { t } = useI18n();
  const [meals, setMeals] = createSignal<any[]>([]);
  const [showModal, setShowModal] = createSignal(false);
  const [selected, setSelected] = createSignal<any>(null);
  const [price, setPrice] = createSignal<number>(0);
  const [effectiveFrom, setEffectiveFrom] = createSignal<string>(
    new Date().toISOString().slice(0, 16),
  );
  const [history, setHistory] = createSignal<any[]>([]);

  const formatLocal = (iso?: string) => formatLocalDateTime(iso);

  const fetchMeals = async () => {
    try {
      const res = await axios.get("/api/meals");
      setMeals(res.data || []);
    } catch (err) {
      console.error("Failed to fetch meals", err);
    }
  };

  const fetchHistory = async (itemId: string) => {
    try {
      const res = await axios.get(`/api/meals/${itemId}/prices`);
      setHistory(res.data || []);
    } catch (err) {
      console.error("Failed to fetch price history", err);
      setHistory([]);
    }
  };

  const openModal = (m: any) => {
    setSelected(m);
    setPrice(m.price || 0);
    setEffectiveFrom(new Date().toISOString().slice(0, 16));
    setShowModal(true);
    fetchHistory(m.item_id);
  };

  const submitPrice = async () => {
    if (!selected()) return;
    try {
      // send explicit UTC ISO string so backend interprets correctly
      await axios.post(`/api/meals/${selected().item_id}/prices`, {
        price: price(),
        effective_from: new Date(effectiveFrom()).toISOString(),
        created_by: "admin",
      });
      setShowModal(false);
      await fetchMeals();
      alert(t("menuPriceUpdated"));
    } catch (err) {
      console.error(err);
      alert(t("failedToUpdatePrice"));
    }
  };

  onMount(() => {
    fetchMeals();
  });

  return (
    <div class="space-y-6">
      <div class="md-card p-6">
        <h3 class="text-xl font-bold mb-4">{t("menuPricing")}</h3>
        <div class="grid gap-4">
          <For each={meals()}>
            {(m: any) => (
              <div class="flex items-center justify-between p-3 bg-white rounded shadow-sm">
                <div>
                  <div class="font-bold">{m.item_name}</div>
                  <div class="text-sm text-slate-500">{m.item_id}</div>
                </div>
                <div class="flex items-center gap-4">
                  <div class="text-lg font-bold">{m.price?.toFixed(2)}</div>
                  <button class="btn btn-primary" onClick={() => openModal(m)}>
                    {t("updatePrice")}
                  </button>
                </div>
              </div>
            )}
          </For>
        </div>
      </div>

      {showModal() && (
        <div class="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div class="bg-white rounded p-6 w-full max-w-md">
            <h4 class="text-lg font-bold mb-4">
              {t("updatePriceFor")} {selected()?.item_name}
            </h4>
            <div class="space-y-3">
              <div>
                <label class="block text-sm text-slate-600">{t("price")}</label>
                <input
                  type="number"
                  step="0.01"
                  class="input"
                  value={price()}
                  onInput={(e: any) => setPrice(parseFloat(e.target.value))}
                />
              </div>
              <div>
                <label class="block text-sm text-slate-600">
                  {t("effectiveFrom")}
                </label>
                <input
                  type="datetime-local"
                  class="input"
                  value={effectiveFrom()}
                  onInput={(e: any) => setEffectiveFrom(e.target.value)}
                />
              </div>
              <div class="flex justify-end gap-3 mt-4">
                <button class="btn" onClick={() => setShowModal(false)}>
                  {t("cancel")}
                </button>
                <button class="btn btn-primary" onClick={submitPrice}>
                  {t("save")}
                </button>
              </div>

              {/* Price history shown in same modal */}
              <div class="mt-6">
                <h5 class="text-sm font-medium mb-2">{t("priceHistory")}</h5>
                <div class="space-y-3 max-h-48 overflow-auto">
                  <For each={history()}>
                    {(h: any) => (
                      <div class="flex items-center justify-between p-2 border-b">
                        <div>
                          <div class="font-medium">₹{h.price?.toFixed(2)}</div>
                          <div class="text-xs text-slate-500">
                            {formatLocal(h.effective_from)} (local)
                          </div>
                        </div>
                        <div class="text-xs text-slate-400">
                          {formatLocal(h.created_at)}
                        </div>
                      </div>
                    )}
                  </For>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default MenuPricing;
