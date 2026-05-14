import { createSignal, onMount, For, createEffect } from 'solid-js';
import axios from 'axios';
import { Expense } from '../types';
import { useI18n } from '../i18n';
import {
    Plus,
    Trash2,
    Edit2,
    X,
    Search,
    Calendar,
    IndianRupee,
    FileText,
    Check
} from 'lucide-solid';

const Expenses = () => {
    const { t } = useI18n();
    const [expenses, setExpenses] = createSignal<Expense[]>([]);
    const [startDate, setStartDate] = createSignal(new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString().split('T')[0]);
    const [endDate, setEndDate] = createSignal(new Date().toISOString().split('T')[0]);
    const [isSubmitting, setIsSubmitting] = createSignal(false);
    const [editingExpense, setEditingExpense] = createSignal<Expense | null>(null);
    const [showAddModal, setShowAddModal] = createSignal(false);

    // Form Signals
    const [formDate, setFormDate] = createSignal(new Date().toISOString().split('T')[0]);
    const [formReason, setFormReason] = createSignal('');
    const [formAmount, setFormAmount] = createSignal('');

    const fetchExpenses = async () => {
        try {
            const res = await axios.get(`/api/expenses?start_date=${startDate()}&end_date=${endDate()}`);
            setExpenses(res.data || []);
        } catch (error) {
            console.error('Failed to fetch expenses:', error);
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
                amount: parseFloat(formAmount())
            };

            if (editingExpense()) {
                await axios.put(`/api/expenses/${editingExpense()?.expense_id}`, payload);
            } else {
                await axios.post('/api/expenses', payload);
            }

            // Reset and refresh
            setFormReason('');
            setFormAmount('');
            setEditingExpense(null);
            setShowAddModal(false);
            await fetchExpenses();
        } catch (error) {
            console.error('Failed to save expense:', error);
            alert('Failed to save expense');
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm('Are you sure you want to delete this expense?')) return;
        try {
            await axios.delete(`/api/expenses/${id}`);
            await fetchExpenses();
        } catch (error) {
            console.error('Failed to delete expense:', error);
            alert('Failed to delete expense');
        }
    };

    const openEditModal = (exp: Expense) => {
        setEditingExpense(exp);
        setFormDate(exp.expense_date.split('T')[0]);
        setFormReason(exp.reason);
        setFormAmount(exp.amount.toString());
        setShowAddModal(true);
    };

    const openAddModal = () => {
        setEditingExpense(null);
        setFormDate(new Date().toISOString().split('T')[0]);
        setFormReason('');
        setFormAmount('');
        setShowAddModal(true);
    };

    return (
        <div class="space-y-12 pb-12">
            <header class="flex flex-col md:flex-row md:items-center justify-between gap-6">
                <div>
                    <h2 class="text-4xl font-black text-[var(--md-sys-color-primary)] tracking-tight">{t('expenses')}</h2>
                    <p class="text-[var(--md-sys-color-on-surface-variant)] mt-2 text-lg">{t('manageCosts')}</p>
                </div>
                <button
                    onClick={openAddModal}
                    class="btn-primary flex items-center gap-2 self-start h-[56px] !rounded-2xl"
                >
                    <Plus size={24} />
                    <span class="text-lg font-bold">{t('addExpense')}</span>
                </button>
            </header>

            {/* Filters */}
            <div class="md-card p-6 relative overflow-hidden">
                <div class="absolute top-0 right-0 w-32 h-32 bg-[var(--md-sys-color-secondary)] opacity-5 rounded-full blur-2xl -translate-y-1/2 translate-x-1/2 pointer-events-none"></div>

                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 items-end">
                    <div class="space-y-1">
                        <label class="text-xs font-bold text-[var(--md-sys-color-primary)] ml-4 mb-1 block tracking-wider uppercase flex items-center gap-2">
                            <Calendar size={12} /> {t('startDate')}
                        </label>
                        <input
                            type="date"
                            value={startDate()}
                            onInput={(e) => {
                                const val = e.currentTarget.value;
                                if (val) setStartDate(val);
                            }}
                            class="input-filled"
                        />
                    </div>
                    <div class="space-y-1">
                        <label class="text-xs font-bold text-[var(--md-sys-color-primary)] ml-4 mb-1 block tracking-wider uppercase flex items-center gap-2">
                            <Calendar size={12} /> {t('endDate')}
                        </label>
                        <input
                            type="date"
                            value={endDate()}
                            onInput={(e) => {
                                const val = e.currentTarget.value;
                                if (val) setEndDate(val);
                            }}
                            class="input-filled"
                        />
                    </div>
                    <div class="lg:col-span-2 flex items-center justify-end h-[56px]">
                        <div class="bg-[var(--md-sys-color-secondary-container)] text-[var(--md-sys-color-on-secondary-container)] px-6 py-3 rounded-2xl flex items-center gap-3 shadow-inner">
                            <IndianRupee size={18} class="opacity-70" />
                            <span class="text-sm font-bold tracking-wide uppercase opacity-70">{t('total')}:</span>
                            <span class="text-2xl font-black">₹{expenses().reduce((sum, e) => sum + e.amount, 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
                        </div>
                    </div>
                </div>
            </div>

            {/* Expenses list */}
            <div class="md-card overflow-hidden slide-in-from-bottom animate-in duration-700 delay-100">
                <div class="overflow-x-auto">
                    <table class="w-full text-left">
                        <thead class="bg-[var(--md-sys-color-surface-container-highest)] border-b border-[var(--md-sys-color-outline-variant)]">
                            <tr>
                                <th class="p-4 font-bold text-[var(--md-sys-color-on-surface-variant)]">{t('date')}</th>
                                <th class="p-4 font-bold text-[var(--md-sys-color-on-surface-variant)]">{t('reason')}</th>
                                <th class="p-4 font-bold text-[var(--md-sys-color-on-surface-variant)] text-right">{t('amount')}</th>
                                <th class="p-4 font-bold text-[var(--md-sys-color-on-surface-variant)] text-center">{t('actions')}</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-[var(--md-sys-color-outline-variant)]">
                            <For each={expenses()}>
                                {(expense) => (
                                    <tr class="hover:bg-[var(--md-sys-color-surface-container-high)] transition-colors group">
                                        <td class="p-4 text-[var(--md-sys-color-on-surface)]">
                                            {new Date(expense.expense_date).toLocaleDateString()}
                                        </td>
                                        <td class="p-4">
                                            <div class="font-medium text-[var(--md-sys-color-on-surface)]">{expense.reason}</div>
                                            <div class="text-xs text-[var(--md-sys-color-on-surface-variant)] opacity-70">
                                                {t('added')} {new Date(expense.created_at).toLocaleString()}
                                            </div>
                                        </td>
                                        <td class="p-4 text-right font-bold text-[var(--md-sys-color-on-surface)] text-lg">
                                            ₹{expense.amount.toLocaleString('en-IN')}
                                        </td>
                                        <td class="p-4">
                                            <div class="flex items-center justify-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                                <button
                                                    onClick={() => openEditModal(expense)}
                                                    class="w-10 h-10 rounded-full flex items-center justify-center hover:bg-[var(--md-sys-color-secondary-container)] text-[var(--md-sys-color-on-surface-variant)] transition-all"
                                                    title="Edit"
                                                >
                                                    <Edit2 size={18} />
                                                </button>
                                                <button
                                                    onClick={() => handleDelete(expense.expense_id)}
                                                    class="w-10 h-10 rounded-full flex items-center justify-center hover:bg-[var(--md-sys-color-error-container)] text-[var(--md-sys-color-error)] transition-all"
                                                    title="Delete"
                                                >
                                                    <Trash2 size={18} />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                )}
                            </For>
                            {expenses().length === 0 && (
                                <tr>
                                    <td colspan="4" class="p-12 text-center text-[var(--md-sys-color-on-surface-variant)] opacity-60">
                                        <div class="flex flex-col items-center gap-2">
                                            <FileText size={48} stroke-width={1} />
                                            <p>{t('noExpenses')}</p>
                                        </div>
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Add/Edit Modal */}
            {showAddModal() && (
                <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm animate-in fade-in duration-300">
                    <div class="md-card w-full max-w-md shadow-2xl zoom-in animate-in duration-300">
                        <div class="p-6 border-b border-[var(--md-sys-color-outline-variant)] flex items-center justify-between">
                            <h3 class="text-xl font-bold text-[var(--md-sys-color-on-surface)]">
                                {editingExpense() ? t('editExpense') : t('addExpense')}
                            </h3>
                            <button
                                onClick={() => setShowAddModal(false)}
                                class="w-10 h-10 rounded-full flex items-center justify-center hover:bg-[var(--md-sys-color-surface-container-highest)] text-[var(--md-sys-color-on-surface-variant)] transition-all"
                            >
                                <X size={20} />
                            </button>
                        </div>

                        <form onSubmit={handleSubmit} class="p-6 space-y-6">
                            <div class="space-y-1">
                                <label class="text-xs font-bold text-[var(--md-sys-color-primary)] ml-4 mb-1 block tracking-wider uppercase flex items-center gap-2">
                                    <Calendar size={12} /> {t('date')}
                                </label>
                                <input
                                    type="date"
                                    required
                                    value={formDate()}
                                    onInput={(e) => {
                                        const val = e.currentTarget.value;
                                        if (val) setFormDate(val);
                                    }}
                                    class="input-filled"
                                />
                            </div>

                            <div class="space-y-1">
                                <label class="text-xs font-bold text-[var(--md-sys-color-primary)] ml-4 mb-1 block tracking-wider uppercase flex items-center gap-2">
                                    <FileText size={12} /> {t('reasonNote')}
                                </label>
                                <textarea
                                    required
                                    placeholder={t('explainExpense')}
                                    value={formReason()}
                                    onInput={(e) => setFormReason(e.currentTarget.value)}
                                    class="input-filled min-h-[120px] py-4 leading-relaxed"
                                />
                            </div>

                            <div class="space-y-1">
                                <label class="text-xs font-bold text-[var(--md-sys-color-primary)] ml-4 mb-1 block tracking-wider uppercase flex items-center gap-2">
                                    <IndianRupee size={12} /> {t('amount')}
                                </label>
                                <input
                                    type="number"
                                    required
                                    step="0.01"
                                    placeholder="0.00"
                                    value={formAmount()}
                                    onInput={(e) => setFormAmount(e.currentTarget.value)}
                                    class="input-filled font-black text-2xl !border-b-[var(--md-sys-color-tertiary)]"
                                />
                            </div>

                            <div class="pt-4">
                                <button
                                    type="submit"
                                    disabled={isSubmitting()}
                                    class="btn-primary w-full h-16 rounded-2xl font-black text-xl flex items-center justify-center gap-3 shadow-xl shadow-[var(--md-sys-color-primary)]/20"
                                >
                                    {isSubmitting() ? (
                                        <div class="w-6 h-6 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                                    ) : (
                                        <>
                                            <Check size={20} stroke-width={3} />
                                            {editingExpense() ? t('saveChanges') : t('addExpense')}
                                        </>
                                    )}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Expenses;
