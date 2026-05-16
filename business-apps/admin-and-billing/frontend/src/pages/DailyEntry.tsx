import axios from 'axios';
import { Check, ChefHat, Moon, Salad, SquarePen, Sun, Trash2, Utensils, X } from 'lucide-solid';
import { For, createEffect, createSignal, onMount } from 'solid-js';
import { DailyLog, User } from '../types';
import { useI18n } from '../i18n';

import { globalUsers, globalUserTrie, loadUsers, updateUserBalance } from '../store/userStore';

const DailyEntry = () => {
    const { t } = useI18n();
    const [prices, setPrices] = createSignal<Record<string, number>>({});
    const [logs, setLogs] = createSignal<DailyLog[]>([]);
    const [selectedUser, setSelectedUser] = createSignal<string>('');
    const [searchQuery, setSearchQuery] = createSignal('');
    const [suggestions, setSuggestions] = createSignal<User[]>([]);
    const [showSuggestions, setShowSuggestions] = createSignal(false);
    const [date, setDate] = createSignal(new Date().toISOString().split('T')[0]);
    const [mealType, setMealType] = createSignal<'lunch' | 'dinner'>('lunch');
    const [mealCategory, setMealCategory] = createSignal<'standard' | 'special' | 'none'>('standard');
    const [specialDish, setSpecialDish] = createSignal('');
    const [extraRice, setExtraRice] = createSignal(0);
    const [extraRoti, setExtraRoti] = createSignal(0);
    const [extraChicken, setExtraChicken] = createSignal(0);
    const [extraFish, setExtraFish] = createSignal(0);
    const [extraEgg, setExtraEgg] = createSignal(0);
    const [extraVegetable, setExtraVegetable] = createSignal(0);
    const [isSubmitting, setIsSubmitting] = createSignal(false);
    const [successMsg, setSuccessMsg] = createSignal(false);
    const [editingLog, setEditingLog] = createSignal<DailyLog | null>(null);



    const handleSearchInput = (e: any) => {
        const val = e.currentTarget.value;
        setSearchQuery(val);
        
        if (selectedUser()) {
            setSelectedUser(''); 
        }
        
        if (val.trim().length >= 3) {
            const results = globalUserTrie().search(val.trim(), 5);
            setSuggestions(results);
            setShowSuggestions(true);
        } else {
            setSuggestions([]);
            setShowSuggestions(false);
        }
    };

    const handleSelectUser = (u: User) => {
        setSelectedUser(u.user_id.toString());
        setSearchQuery(u.name);
        setShowSuggestions(false);
    };

    const handleSearchBlur = () => {
        setTimeout(() => setShowSuggestions(false), 200);
    };

    const fetchPrices = async () => {
        try {
            const res = await axios.get('/api/meals');
            const priceMap: Record<string, number> = {};
            res.data?.forEach((item: any) => {
                priceMap[item.item_id] = item.price;
            });
            setPrices(priceMap);
        } catch (error) {
            console.error('Failed to fetch prices:', error);
        }
    };

    onMount(() => {
        loadUsers();
        fetchPrices();
    });

    const fetchLogs = async () => {
        try {
            const userId = selectedUser();
            const res = await axios.get(`/api/daily-entry?date=${date()}${userId ? `&user_id=${userId}` : ''}`);
            setLogs(res.data || []);
        } catch (error) {
            console.error('Failed to fetch logs:', error);
        }
    };

    // Construct a derived signal or effect to fetch logs when date or selectedUser changes
    createEffect(() => {
        // Track date and selectedUser
        date();
        selectedUser();
        fetchLogs();
    });

    const handleDelete = async (logId: number) => {
        const logToDelete = logs().find(l => l.log_id === logId);
        if (!logToDelete) return;
        if (!confirm('Are you sure you want to delete this entry? This will refund the cost to the user\'s wallet.')) return;

        try {
            const res = await axios.delete(`/api/daily-entry/${logId}`);
            setSuccessMsg(true); // Reuse success msg or create a new one
            setTimeout(() => setSuccessMsg(false), 3000);
            await fetchLogs();
            updateUserBalance(logToDelete.user_id, res.data.new_balance); // Update wallet balance
        } catch (error) {
            console.error('Failed to delete log:', error);
            alert('Failed to delete entry');
        }
    };

    const handleSubmit = async (e: Event) => {
        e.preventDefault();
        if (!selectedUser() || !date()) return;

        setIsSubmitting(true);
        try {
            const res = await axios.post('/api/daily-entry', {
                user_id: parseInt(selectedUser()),
                log_date: new Date(date()).toISOString(),
                meal_type: mealType(),
                has_main_meal: mealCategory() !== 'none',
                is_special: mealCategory() === 'special',
                special_dish_name: mealCategory() === 'special' ? specialDish() : '',
                extra_rice_qty: extraRice(),
                extra_roti_qty: extraRoti(),
                extra_chicken_qty: extraChicken(),
                extra_fish_qty: extraFish(),
                extra_egg_qty: extraEgg(),
                extra_vegetable_qty: extraVegetable()
            });
            setSuccessMsg(true);
            setTimeout(() => setSuccessMsg(false), 3000);
            updateUserBalance(parseInt(selectedUser()), res.data.new_balance);
            await fetchLogs();

            // Reset some fields
            setMealCategory('standard');
            setSpecialDish('');
            setExtraRice(0);
            setExtraRoti(0);
            setExtraChicken(0);
            setExtraFish(0);
            setExtraEgg(0);
            setExtraVegetable(0);
        } catch (err) {
            alert('Failed to record entry');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div class="max-w-xl mx-auto py-8 animate-in slide-in-from-bottom">
            <header class="mb-8 text-center">
                <h2 class="text-4xl font-black text-[var(--md-sys-color-primary)] tracking-tight">{t('dailyEntry')}</h2>
                <p class="text-[var(--md-sys-color-on-surface-variant)] mt-2 text-lg">{t('recordConsumption')}</p>
            </header>

            <form onSubmit={handleSubmit} class="md-card flex flex-col gap-6 shadow-2xl relative overflow-hidden">
                {/* Decorative background element */}
                <div class="absolute top-0 right-0 w-64 h-64 bg-[var(--md-sys-color-primary)] opacity-5 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2 pointer-events-none"></div>

                {/* Date & Shift Group */}
                <div class="flex gap-4">
                    <div class="flex-1">
                        <label class="text-xs font-bold text-[var(--md-sys-color-primary)] ml-4 mb-1 block tracking-wider uppercase">{t('date')}</label>
                        <input
                            type="date"
                            class="input-filled"
                            value={date()}
                            onInput={e => {
                                const val = e.currentTarget.value;
                                if (val) setDate(val);
                            }}
                        />
                    </div>
                    <div class="flex-1">
                        <label class="text-xs font-bold text-[var(--md-sys-color-primary)] ml-4 mb-1 block tracking-wider uppercase">{t('shift')}</label>
                        <div class="bg-[var(--md-sys-color-surface-container-highest)] rounded-full p-1 flex h-[56px] items-center">
                            <button
                                type="button"
                                onClick={() => setMealType('lunch')}
                                class={`flex-1 h-full rounded-full text-sm font-bold flex items-center justify-center gap-2 transition-all ${mealType() === 'lunch' ? 'bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] shadow-md' : 'text-[var(--md-sys-color-on-surface-variant)] hover:bg-white/5'}`}
                            >
                                <Sun size={16} /> {t('lunch')}
                            </button>
                            <button
                                type="button"
                                onClick={() => setMealType('dinner')}
                                class={`flex-1 h-full rounded-full text-sm font-bold flex items-center justify-center gap-2 transition-all ${mealType() === 'dinner' ? 'bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] shadow-md' : 'text-[var(--md-sys-color-on-surface-variant)] hover:bg-white/5'}`}
                            >
                                <Moon size={16} /> {t('dinner')}
                            </button>
                        </div>
                    </div>
                </div>

                {/* Customer Select */}
                <div>
                    <label class="text-xs font-bold text-[var(--md-sys-color-primary)] ml-4 mb-1 block tracking-wider uppercase">{t('customer')}</label>
                    <div class="relative">
                        <input
                            type="text"
                            class="input-filled w-full"
                            placeholder={t('selectCustomer') || "Search Customer..."}
                            value={searchQuery()}
                            onInput={handleSearchInput}
                            onFocus={() => { if (searchQuery().trim().length >= 3) setShowSuggestions(true); }}
                            onBlur={handleSearchBlur}
                            required={!selectedUser()}
                        />
                        {showSuggestions() && (
                            <ul class="absolute z-50 top-[calc(100%+4px)] left-0 right-0 bg-[var(--md-sys-color-surface-container)] rounded-2xl shadow-xl border border-[var(--md-sys-color-outline-variant)] overflow-hidden">
                                <For each={suggestions()}>
                                    {(user) => (
                                        <li 
                                            class="px-5 py-3 cursor-pointer hover:bg-[var(--md-sys-color-surface-container-high)] border-b border-[var(--md-sys-color-outline-variant)] last:border-0 flex justify-between items-center transition-colors"
                                            onClick={() => handleSelectUser(user)}
                                        >
                                            <span class="font-bold text-[var(--md-sys-color-on-surface)]">{user.name}</span>
                                            <span class="text-xs font-bold text-[var(--md-sys-color-on-surface-variant)] bg-[var(--md-sys-color-surface-container-highest)] px-2.5 py-1 rounded-lg tracking-wide">
                                                {user.role === 'admin' ? 'Admin' : 'User'} • ₹{user.balance.toFixed(0)}
                                            </span>
                                        </li>
                                    )}
                                </For>
                                {suggestions().length === 0 && (
                                    <li class="px-5 py-6 text-sm text-[var(--md-sys-color-on-surface-variant)] text-center font-medium">
                                        No matching customers found
                                    </li>
                                )}
                            </ul>
                        )}
                        <div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-[var(--md-sys-color-on-surface-variant)]">
                            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
                        </div>
                    </div>
                </div>

                {/* Meal Selection Cards */}
                <div>
                    <label class="text-xs font-bold text-[var(--md-sys-color-primary)] ml-4 mb-3 block tracking-wider uppercase">{t('mealType')}</label>
                    <div class="grid grid-cols-3 gap-3">
                        {/* Standard */}
                        <label class={`relative flex flex-col items-center p-4 rounded-[20px] cursor-pointer transition-all duration-300 border-2 ${mealCategory() === 'standard'
                            ? 'bg-[var(--md-sys-color-secondary-container)] border-[var(--md-sys-color-secondary-container)] text-[var(--md-sys-color-on-secondary-container)]'
                            : 'border-[var(--md-sys-color-outline-variant)] text-[var(--md-sys-color-on-surface-variant)] hover:border-[var(--md-sys-color-outline)] hover:bg-[var(--md-sys-color-surface-container-high)]'
                            }`}>
                            <input type="radio" name="cat" class="hidden" checked={mealCategory() === 'standard'} onChange={() => setMealCategory('standard')} />
                            <Utensils size={24} class="mb-2" />
                            <span class="text-sm font-bold">{t('standard')}</span>
                            <span class="text-[10px] opacity-80 mt-1">₹{prices()['standard'] ?? 52.5}</span>
                        </label>

                        {/* Special */}
                        <label class={`relative flex flex-col items-center p-4 rounded-[20px] cursor-pointer transition-all duration-300 border-2 ${mealCategory() === 'special'
                            ? 'bg-[var(--md-sys-color-tertiary-container)] border-[var(--md-sys-color-tertiary-container)] text-[var(--md-sys-color-on-tertiary-container)]'
                            : 'border-[var(--md-sys-color-outline-variant)] text-[var(--md-sys-color-on-surface-variant)] hover:border-[var(--md-sys-color-outline)] hover:bg-[var(--md-sys-color-surface-container-high)]'
                            }`}>
                            <input type="radio" name="cat" class="hidden" checked={mealCategory() === 'special'} onChange={() => setMealCategory('special')} />
                            <ChefHat size={24} class="mb-2" />
                            <span class="text-sm font-bold">Special</span>
                            <span class="text-[10px] opacity-80 mt-1">₹{prices()['special'] ?? 120}</span>
                        </label>

                        {/* None */}
                        <label class={`relative flex flex-col items-center p-4 rounded-[20px] cursor-pointer transition-all duration-300 border-2 ${mealCategory() === 'none'
                            ? 'bg-[var(--md-sys-color-surface-variant)] border-[var(--md-sys-color-on-surface)] text-[var(--md-sys-color-on-surface)]'
                            : 'border-[var(--md-sys-color-outline-variant)] text-[var(--md-sys-color-on-surface-variant)] hover:border-[var(--md-sys-color-outline)] hover:bg-[var(--md-sys-color-surface-container-high)]'
                            }`}>
                            <input type="radio" name="cat" class="hidden" checked={mealCategory() === 'none'} onChange={() => setMealCategory('none')} />
                            <Salad size={24} class="mb-2" />
                            <span class="text-sm font-bold">{t('aLaCarte')}</span>
                            <span class="text-[10px] opacity-80 mt-1">{t('extrasOnly')}</span>
                        </label>
                    </div>
                </div>

                {/* Special Dish Input */}
                {mealCategory() === 'special' && (
                    <div class="animate-in fade-in slide-in-from-bottom duration-300">
                        <label class="text-xs font-bold text-[var(--md-sys-color-tertiary)] ml-4 mb-1 block tracking-wider uppercase">{t('dishName')}</label>
                        <input
                            class="input-filled !border-b-[var(--md-sys-color-tertiary)]" // Override border color for tertiary feel
                            placeholder={t('egMutton')}
                            required
                            value={specialDish()}
                            onInput={e => setSpecialDish(e.currentTarget.value)}
                        />
                    </div>
                )}

                {/* Extras */}
                <div class="pt-4 border-t border-[var(--md-sys-color-outline-variant)]">
                    <h4 class="text-sm font-bold text-[var(--md-sys-color-on-surface-variant)] mb-4 uppercase tracking-wider ml-4">{t('extras')}</h4>
                    <div class="grid grid-cols-2 gap-4">
                        {/* Rice */}
                        <div class="bg-[var(--md-sys-color-surface-container-high)] p-4 rounded-2xl flex flex-col items-center">
                            <span class="text-xs font-medium text-[var(--md-sys-color-on-surface-variant)] mb-2">{t('rice')} (₹{prices()['rice'] ?? 10})</span>
                            <div class="flex items-center gap-4">
                                <button type="button" onClick={() => setExtraRice(Math.max(0, extraRice() - 1))} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-surface-container-highest)] hover:bg-[var(--md-sys-color-primary-container)] hover:text-[var(--md-sys-color-on-primary-container)] transition-colors flex items-center justify-center font-bold text-xl">-</button>
                                <span class="text-xl font-bold w-6 text-center">{extraRice()}</span>
                                <button type="button" onClick={() => setExtraRice(extraRice() + 1)} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] hover:opacity-90 transition-colors flex items-center justify-center font-bold text-xl">+</button>
                            </div>
                        </div>
                        {/* Roti */}
                        <div class="bg-[var(--md-sys-color-surface-container-high)] p-4 rounded-2xl flex flex-col items-center">
                            <span class="text-xs font-medium text-[var(--md-sys-color-on-surface-variant)] mb-2">{t('roti')} (₹{prices()['roti'] ?? 4})</span>
                            <div class="flex items-center gap-4">
                                <button type="button" onClick={() => setExtraRoti(Math.max(0, extraRoti() - 1))} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-surface-container-highest)] hover:bg-[var(--md-sys-color-primary-container)] hover:text-[var(--md-sys-color-on-primary-container)] transition-colors flex items-center justify-center font-bold text-xl">-</button>
                                <span class="text-xl font-bold w-6 text-center">{extraRoti()}</span>
                                <button type="button" onClick={() => setExtraRoti(extraRoti() + 1)} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] hover:opacity-90 transition-colors flex items-center justify-center font-bold text-xl">+</button>
                            </div>
                        </div>
                        {/* Chicken */}
                        <div class="bg-[var(--md-sys-color-surface-container-high)] p-4 rounded-2xl flex flex-col items-center">
                            <span class="text-xs font-medium text-[var(--md-sys-color-on-surface-variant)] mb-2">{t('chicken')} (₹{prices()['chicken'] ?? 30})</span>
                            <div class="flex items-center gap-4">
                                <button type="button" onClick={() => setExtraChicken(Math.max(0, extraChicken() - 1))} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-surface-container-highest)] hover:bg-[var(--md-sys-color-primary-container)] hover:text-[var(--md-sys-color-on-primary-container)] transition-colors flex items-center justify-center font-bold text-xl">-</button>
                                <span class="text-xl font-bold w-6 text-center">{extraChicken()}</span>
                                <button type="button" onClick={() => setExtraChicken(extraChicken() + 1)} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] hover:opacity-90 transition-colors flex items-center justify-center font-bold text-xl">+</button>
                            </div>
                        </div>
                        {/* Fish */}
                        <div class="bg-[var(--md-sys-color-surface-container-high)] p-4 rounded-2xl flex flex-col items-center">
                            <span class="text-xs font-medium text-[var(--md-sys-color-on-surface-variant)] mb-2">{t('fish')} (₹{prices()['fish'] ?? 20})</span>
                            <div class="flex items-center gap-4">
                                <button type="button" onClick={() => setExtraFish(Math.max(0, extraFish() - 1))} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-surface-container-highest)] hover:bg-[var(--md-sys-color-primary-container)] hover:text-[var(--md-sys-color-on-primary-container)] transition-colors flex items-center justify-center font-bold text-xl">-</button>
                                <span class="text-xl font-bold w-6 text-center">{extraFish()}</span>
                                <button type="button" onClick={() => setExtraFish(extraFish() + 1)} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] hover:opacity-90 transition-colors flex items-center justify-center font-bold text-xl">+</button>
                            </div>
                        </div>
                        {/* Egg */}
                        <div class="bg-[var(--md-sys-color-surface-container-high)] p-4 rounded-2xl flex flex-col items-center">
                            <span class="text-xs font-medium text-[var(--md-sys-color-on-surface-variant)] mb-2">{t('egg')} (₹{prices()['egg'] ?? 10})</span>
                            <div class="flex items-center gap-4">
                                <button type="button" onClick={() => setExtraEgg(Math.max(0, extraEgg() - 1))} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-surface-container-highest)] hover:bg-[var(--md-sys-color-primary-container)] hover:text-[var(--md-sys-color-on-primary-container)] transition-colors flex items-center justify-center font-bold text-xl">-</button>
                                <span class="text-xl font-bold w-6 text-center">{extraEgg()}</span>
                                <button type="button" onClick={() => setExtraEgg(extraEgg() + 1)} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] hover:opacity-90 transition-colors flex items-center justify-center font-bold text-xl">+</button>
                            </div>
                        </div>
                        {/* Vegetable */}
                        <div class="bg-[var(--md-sys-color-surface-container-high)] p-4 rounded-2xl flex flex-col items-center">
                            <span class="text-xs font-medium text-[var(--md-sys-color-on-surface-variant)] mb-2">{t('vegetable')} (₹{prices()['vegetable'] ?? 15})</span>
                            <div class="flex items-center gap-4">
                                <button type="button" onClick={() => setExtraVegetable(Math.max(0, extraVegetable() - 1))} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-surface-container-highest)] hover:bg-[var(--md-sys-color-primary-container)] hover:text-[var(--md-sys-color-on-primary-container)] transition-colors flex items-center justify-center font-bold text-xl">-</button>
                                <span class="text-xl font-bold w-6 text-center">{extraVegetable()}</span>
                                <button type="button" onClick={() => setExtraVegetable(extraVegetable() + 1)} class="w-10 h-10 rounded-full bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] hover:opacity-90 transition-colors flex items-center justify-center font-bold text-xl">+</button>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="pt-2">
                    <button
                        type="submit"
                        disabled={isSubmitting()}
                        class={`btn-primary w-full h-14 text-lg rounded-[20px] relative overflow-hidden group ${isSubmitting() ? 'opacity-50' : ''}`}
                    >
                        <div class="absolute inset-0 bg-white/20 translate-y-full group-hover:translate-y-0 transition-transform duration-300"></div>
                        <span class="relative flex items-center gap-2">
                            {isSubmitting() ? t('recording') : <><Check size={20} class="stroke-[3]" /> {t('recordEntry')}</>}
                        </span>
                    </button>
                </div>
            </form>

            {/* Recent Entries Table */}
            <div class="md-card mt-4 p-6 slide-in-from-bottom animate-in duration-700 delay-100">
                <h3 class="text-xl font-bold text-[var(--md-sys-color-primary)] mb-4 flex items-center gap-2">
                    <span class="bg-[var(--md-sys-color-primary-container)] text-[var(--md-sys-color-on-primary-container)] px-2 py-1 rounded-lg text-sm">
                        {date()}
                    </span>
                    {t('entries')}
                </h3>

                <div class="overflow-x-auto">
                    <table class="w-full text-left text-sm">
                        <thead class="bg-[var(--md-sys-color-surface-container-highest)] text-[var(--md-sys-color-on-surface-variant)]">
                            <tr>
                                <th class="p-3 rounded-l-xl">{t('customer')}</th>
                                <th class="p-3">{t('meal')}</th>
                                <th class="p-3">{t('details')}</th>
                                <th class="p-3 text-right">{t('cost')}</th>
                                <th class="p-3 rounded-r-xl text-center">{t('actions')}</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-[var(--md-sys-color-outline-variant)]">
                            <For each={logs()}>
                                {(log) => (
                                    <tr class="hover:bg-[var(--md-sys-color-surface-container-high)] transition-colors">
                                        <td class="p-3 font-medium">{log.user_name || globalUsers().find(u => u.user_id === log.user_id)?.name || log.user_id}</td>
                                        <td class="p-3 capitalize">
                                            <span class={`px-2 py-1 rounded-md text-xs font-bold ${log.meal_type === 'lunch' ? 'bg-amber-100 text-amber-800' : 'bg-indigo-100 text-indigo-800'}`}>
                                                {log.meal_type}
                                            </span>
                                        </td>
                                        <td class="p-3 text-[var(--md-sys-color-on-surface-variant)]">
                                            {log.has_main_meal ? (log.is_special ? log.special_dish_name : t('standard')) : t('extrasOnly')}
                                            {(log.extra_rice_qty > 0 || log.extra_roti_qty > 0 || log.extra_chicken_qty > 0 || log.extra_fish_qty > 0 || log.extra_egg_qty > 0 || log.extra_vegetable_qty > 0) && (
                                                <span class="text-xs ml-2 opacity-70">
                                                    (Rice: {log.extra_rice_qty}, Roti: {log.extra_roti_qty}, Chicken: {log.extra_chicken_qty}, Fish: {log.extra_fish_qty}, Egg: {log.extra_egg_qty}, Vegetable: {log.extra_vegetable_qty})
                                                </span>
                                            )}
                                        </td>
                                        <td class="p-3 text-right font-bold">₹{log.total_cost}</td>
                                        <td class="p-3 flex justify-center gap-2">
                                            <button
                                                class="w-8 h-8 rounded-full hover:bg-[var(--md-sys-color-secondary-container)] hover:text-[var(--md-sys-color-on-secondary-container)] flex items-center justify-center transition-colors text-[var(--md-sys-color-on-surface-variant)]"
                                                title="Edit"
                                                onClick={() => setEditingLog(log)}
                                            >
                                                <SquarePen size={16} />
                                            </button>
                                            <button
                                                class="w-8 h-8 rounded-full hover:bg-[var(--md-sys-color-error-container)] hover:text-[var(--md-sys-color-on-error-container)] flex items-center justify-center transition-colors text-[var(--md-sys-color-error)]"
                                                title="Delete"
                                                onClick={() => handleDelete(log.log_id)}
                                            >
                                                <Trash2 size={16} />
                                            </button>
                                        </td>
                                    </tr>
                                )}
                            </For>
                            {logs().length === 0 && (
                                <tr>
                                    <td colspan={5} class="p-8 text-center text-[var(--md-sys-color-outline)]">
                                        {t('noEntries')}
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            <div class={`fixed bottom-8 left-1/2 -translate-x-1/2 z-50 bg-[var(--md-sys-color-tertiary-container)] text-[var(--md-sys-color-on-tertiary-container)] border border-[var(--md-sys-color-tertiary)] px-6 py-4 rounded-full shadow-2xl flex items-center gap-3 transition-all duration-300 ${successMsg() ? 'translate-y-0 opacity-100' : 'translate-y-20 opacity-0'}`}>
                <div class="bg-[var(--md-sys-color-on-tertiary-container)] rounded-full p-1"><Check size={16} class="text-[var(--md-sys-color-tertiary-container)]" /></div>
                <span class="font-bold tracking-wide">{t('mealRecorded')}</span>
            </div>

            {/* Edit Modal */}
            {editingLog() && (
                <EditLogModal
                    log={editingLog()!}
                    onClose={() => setEditingLog(null)}
                    onSuccess={async (newBalance: number) => {
                        setEditingLog(null);
                        await fetchLogs();
                        updateUserBalance(editingLog()!.user_id, newBalance);
                        setSuccessMsg(true);
                        setTimeout(() => setSuccessMsg(false), 3000);
                    }}
                />
            )}
        </div>
    );
};

const EditLogModal = (props: { log: DailyLog; onClose: () => void; onSuccess: (newBalance: number) => void }) => {
    const [mealType, setMealType] = createSignal(props.log.meal_type);
    const [extraRice, setExtraRice] = createSignal(props.log.extra_rice_qty);
    const [extraRoti, setExtraRoti] = createSignal(props.log.extra_roti_qty);
    const [isSubmitting, setIsSubmitting] = createSignal(false);

    // Derived category from props for initial state if possible, or just default to simple edits
    // For simplicity, let's allow editing Shift (Meal Type) and Extras. 
    // Changing standard/special is complex UI, let's stick to extras/shift for now as those are most common corrections. 
    // If they need to change meal Categoy, they might prefer deleting and re-adding, but let's try to support Extras edit.

    const handleSubmit = async (e: Event) => {
        e.preventDefault();
        setIsSubmitting(true);
        try {
            const res = await axios.put(`/api/daily-entry/${props.log.log_id}`, {
                meal_type: mealType(),
                extra_rice_qty: extraRice(),
                extra_roti_qty: extraRoti(),
                has_main_meal: props.log.has_main_meal,
                is_special: props.log.is_special
                // Pass other fields as is or undefined if backend handles partial updates
            });
            props.onSuccess(res.data.new_balance);
        } catch (err) {
            alert('Failed to update entry');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div class="fixed inset-0 z-[60] grid place-items-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-300">
            <div class="glass w-full max-w-md p-6 shadow-2xl animate-in zoom-in duration-300 bg-[var(--md-sys-color-surface)] relative rounded-3xl">
                <button onClick={props.onClose} class="absolute top-4 right-4 text-[var(--md-sys-color-on-surface-variant)] hover:text-[var(--md-sys-color-on-surface)]">
                    <X size={24} />
                </button>
                <h3 class="text-2xl font-bold text-[var(--md-sys-color-on-surface)] mb-6">Edit Entry</h3>

                <form onSubmit={handleSubmit} class="space-y-6">
                    <div>
                        <label class="block text-sm font-bold text-[var(--md-sys-color-primary)] mb-2 uppercase tracking-wider">Shift</label>
                        <div class="flex gap-2">
                            <button
                                type="button"
                                onClick={() => setMealType('lunch')}
                                class={`flex-1 py-3 rounded-full text-sm font-bold flex items-center justify-center gap-2 transition-all ${mealType() === 'lunch' ? 'bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] shadow-md' : 'bg-[var(--md-sys-color-surface-container-highest)] text-[var(--md-sys-color-on-surface-variant)]'}`}
                            >
                                <Sun size={16} /> Lunch
                            </button>
                            <button
                                type="button"
                                onClick={() => setMealType('dinner')}
                                class={`flex-1 py-3 rounded-full text-sm font-bold flex items-center justify-center gap-2 transition-all ${mealType() === 'dinner' ? 'bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] shadow-md' : 'bg-[var(--md-sys-color-surface-container-highest)] text-[var(--md-sys-color-on-surface-variant)]'}`}
                            >
                                <Moon size={16} /> Dinner
                            </button>
                        </div>
                    </div>

                    <div>
                        <label class="block text-sm font-bold text-[var(--md-sys-color-primary)] mb-2 uppercase tracking-wider">Extras</label>
                        <div class="grid grid-cols-2 gap-4">
                            <div class="bg-[var(--md-sys-color-surface-container-high)] p-3 rounded-xl flex flex-col items-center">
                                <span class="text-xs font-medium text-[var(--md-sys-color-on-surface-variant)] mb-2">Rice</span>
                                <div class="flex items-center gap-3">
                                    <button type="button" onClick={() => setExtraRice(Math.max(0, extraRice() - 1))} class="w-8 h-8 rounded-full bg-[var(--md-sys-color-surface-container-highest)] font-bold">-</button>
                                    <span class="font-bold w-4 text-center">{extraRice()}</span>
                                    <button type="button" onClick={() => setExtraRice(extraRice() + 1)} class="w-8 h-8 rounded-full bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] font-bold">+</button>
                                </div>
                            </div>
                            <div class="bg-[var(--md-sys-color-surface-container-high)] p-3 rounded-xl flex flex-col items-center">
                                <span class="text-xs font-medium text-[var(--md-sys-color-on-surface-variant)] mb-2">Roti</span>
                                <div class="flex items-center gap-3">
                                    <button type="button" onClick={() => setExtraRoti(Math.max(0, extraRoti() - 1))} class="w-8 h-8 rounded-full bg-[var(--md-sys-color-surface-container-highest)] font-bold">-</button>
                                    <span class="font-bold w-4 text-center">{extraRoti()}</span>
                                    <button type="button" onClick={() => setExtraRoti(extraRoti() + 1)} class="w-8 h-8 rounded-full bg-[var(--md-sys-color-primary)] text-[var(--md-sys-color-on-primary)] font-bold">+</button>
                                </div>
                            </div>
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={isSubmitting()}
                        class="btn-primary w-full h-12 rounded-xl font-bold text-lg"
                    >
                        {isSubmitting() ? 'Saving...' : 'Save Changes'}
                    </button>
                </form>
            </div>
        </div>
    );
};

export default DailyEntry;
