import { createSignal, onMount, For, createEffect } from "solid-js";
import axios from "axios";
import { Expense } from "../types";
import { useI18n } from "../i18n";
import { formatLocalDate } from "../utils/format";
import {
  Plus,
  Trash2,
  Edit2,
  X,
  Calendar,
  IndianRupee,
  FileText,
  Check,
} from "lucide-solid";

const Expenses = () => {
  const { t } = useI18n();
  const [expenses, setExpenses] = createSignal<Expense[]>([]);

  // Default to current month
  const getFirstDayOfMonth = () =>
    new Date(new Date().getFullYear(), new Date().getMonth(), 1)
      .toISOString()
      .split("T")[0];
  const getToday = () => new Date().toISOString().split("T")[0];

  const [startDate, setStartDate] = createSignal(getFirstDayOfMonth());
  const [endDate, setEndDate] = createSignal(getToday());

  const [isSubmitting, setIsSubmitting] = createSignal(false);
  const [editingExpense, setEditingExpense] = createSignal<Expense | null>(
    null,
  );
  const [showAddModal, setShowAddModal] = createSignal(false);

  // Form Signals
  const [formDate, setFormDate] = createSignal(getToday());
  const [formReason, setFormReason] = createSignal("");
  const [formAmount, setFormAmount] = createSignal("");

  const fetchExpenses = async () => {
    try {
      const res = await axios.get(
        `/api/expenses?start_date=${startDate()}&end_date=${endDate()}`,
      );
      setExpenses(res.data || []);
    } catch (error) {
      console.error("Failed to fetch expenses:", error);
    }
  };

  onMount(fetchExpenses);

  createEffect(() => {
    startDate();
    endDate();
    fetchExpenses();
  });

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const payload = {
        expense_date: new Date(formDate()).toISOString(),
        reason: formReason(),
        amount: parseFloat(formAmount()),
      };

      if (editingExpense()) {
        await axios.put(
          `/api/expenses/${editingExpense()?.expense_id}`,
          payload,
        );
      } else {
        await axios.post("/api/expenses", payload);
      }

      // Reset and refresh
      setFormReason("");
      setFormAmount("");
      setEditingExpense(null);
      setShowAddModal(false);
      await fetchExpenses();
    } catch (error) {
      console.error("Failed to save expense:", error);
      alert("Failed to save expense");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure you want to delete this expense?")) return;
    try {
      await axios.delete(`/api/expenses/${id}`);
      await fetchExpenses();
    } catch (error) {
      console.error("Failed to delete expense:", error);
      alert("Failed to delete expense");
    }
  };

  const openEditModal = (exp: Expense) => {
    setEditingExpense(exp);
    setFormDate(exp.expense_date.split("T")[0]);
    setFormReason(exp.reason);
    setFormAmount(exp.amount.toString());
    setShowAddModal(true);
  };

  const openAddModal = () => {
    setEditingExpense(null);
    setFormDate(getToday());
    setFormReason("");
    setFormAmount("");
    setShowAddModal(true);
  };

  return (
    <div class="space-y-6 pb-24 animate-in">
      <header>
        <h2 class="text-2xl font-bold text-slate-800">{t("expenses")}</h2>
        <p class="text-slate-500 font-medium text-sm mt-1">
          {t("manageCosts")}
        </p>
      </header>

      <button
        onClick={openAddModal}
        class="btn btn-primary w-full shadow-lg shadow-blue-500/30"
      >
        <Plus size={24} />
        {t("addExpense")}
      </button>

      {/* Total Card */}
      <div class="card bg-red-50 border-red-200 flex flex-col items-center justify-center p-6">
        <span class="text-sm font-bold text-red-700 uppercase tracking-wider mb-1">
          {t("total")} Expenses
        </span>
        <span class="text-4xl font-black text-red-900">
          ₹
          {expenses()
            .reduce((sum, e) => sum + e.amount, 0)
            .toLocaleString("en-IN")}
        </span>
      </div>

      {/* Filters */}
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="block text-xs font-bold text-slate-500 mb-1 pl-1 uppercase tracking-wider">
            {t("startDate")}
          </label>
          <input
            type="date"
            value={startDate()}
            onInput={(e) => setStartDate(e.currentTarget.value)}
            class="input-large !h-12 !text-sm !font-bold"
          />
        </div>
        <div>
          <label class="block text-xs font-bold text-slate-500 mb-1 pl-1 uppercase tracking-wider">
            {t("endDate")}
          </label>
          <input
            type="date"
            value={endDate()}
            onInput={(e) => setEndDate(e.currentTarget.value)}
            class="input-large !h-12 !text-sm !font-bold"
          />
        </div>
      </div>

      {/* Expenses list */}
      <div class="space-y-3 pt-2">
        <For each={expenses()}>
          {(expense) => (
            <div class="card p-4 flex flex-col gap-3 bg-white border border-slate-200">
              <div class="flex justify-between items-start">
                <div>
                  <h4 class="font-bold text-lg text-slate-900">
                    {expense.reason}
                  </h4>
                  <div class="flex items-center gap-1 text-slate-500 text-sm font-medium mt-1">
                    <Calendar size={14} />
                    {formatLocalDate(expense.expense_date)}
                  </div>
                </div>
                <span class="text-2xl font-black text-red-600">
                  ₹{expense.amount.toLocaleString("en-IN")}
                </span>
              </div>

              <div class="flex gap-2 justify-end border-t border-slate-100 pt-3 mt-1">
                <button
                  onClick={() => openEditModal(expense)}
                  class="w-10 h-10 rounded-full bg-slate-100 text-slate-600 flex items-center justify-center hover:bg-slate-200 active:bg-slate-300"
                >
                  <Edit2 size={18} />
                </button>
                <button
                  onClick={() => handleDelete(expense.expense_id)}
                  class="w-10 h-10 rounded-full bg-red-100 text-red-600 flex items-center justify-center hover:bg-red-200 active:bg-red-300"
                >
                  <Trash2 size={18} />
                </button>
              </div>
            </div>
          )}
        </For>
        {expenses().length === 0 && (
          <div class="text-center py-10 text-slate-500 font-medium text-lg bg-white rounded-2xl border border-slate-200 border-dashed">
            {t("noExpenses")}
          </div>
        )}
      </div>

      {/* Add/Edit Modal (Bottom Sheet) */}
      {showAddModal() && (
        <div class="fixed inset-0 z-[60] bg-slate-900/40 backdrop-blur-sm flex items-end sm:items-center justify-center">
          <div class="bg-white w-full max-w-[600px] rounded-t-3xl sm:rounded-3xl p-6 shadow-2xl pb-safe animate-in slide-in-from-bottom max-h-[90vh] overflow-y-auto">
            <div class="flex justify-between items-center mb-6">
              <h3 class="text-2xl font-bold text-slate-900">
                {editingExpense() ? t("editExpense") : t("addExpense")}
              </h3>
              <button
                onClick={() => setShowAddModal(false)}
                class="w-10 h-10 bg-slate-100 rounded-full flex items-center justify-center text-slate-600"
              >
                <X size={24} />
              </button>
            </div>

            <form onSubmit={handleSubmit} class="space-y-4">
              <div>
                <label class="block text-sm font-bold text-slate-700 mb-1">
                  {t("amount")}
                </label>
                <div class="relative">
                  <span class="absolute left-4 top-1/2 -translate-y-1/2 text-2xl font-bold text-slate-400">
                    ₹
                  </span>
                  <input
                    type="number"
                    step="0.01"
                    required
                    placeholder="0"
                    value={formAmount()}
                    onInput={(e) => setFormAmount(e.currentTarget.value)}
                    class="input-large !text-3xl !font-black pl-10 !border-red-300 focus:!border-red-500"
                  />
                </div>
              </div>

              <div>
                <label class="block text-sm font-bold text-slate-700 mb-1">
                  {t("reasonNote")}
                </label>
                <textarea
                  required
                  placeholder={t("explainExpense")}
                  value={formReason()}
                  onInput={(e) => setFormReason(e.currentTarget.value)}
                  class="input-large !h-24 py-3 resize-none"
                />
              </div>

              <div>
                <label class="block text-sm font-bold text-slate-700 mb-1">
                  {t("date")}
                </label>
                <input
                  type="date"
                  required
                  value={formDate()}
                  onInput={(e) => setFormDate(e.currentTarget.value)}
                  class="input-large !h-14"
                />
              </div>

              <button
                type="submit"
                disabled={isSubmitting()}
                class="btn btn-primary w-full mt-6 text-xl shadow-lg shadow-blue-500/30"
              >
                {isSubmitting()
                  ? "Saving..."
                  : editingExpense()
                    ? t("saveChanges")
                    : t("addExpense")}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default Expenses;
