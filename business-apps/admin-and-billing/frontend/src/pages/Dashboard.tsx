import { createSignal, onMount, For } from 'solid-js';
import { A } from '@solidjs/router';
import axios from 'axios';
import { User, DashboardStats } from '../types';
import { Users, TrendingUp, TrendingDown, PlusCircle, Wallet, User as UserIcon } from 'lucide-solid';
import { useI18n } from '../i18n';

const Dashboard = () => {
    const { t } = useI18n();
    const [users, setUsers] = createSignal<User[]>([]);
    const [stats, setStats] = createSignal<DashboardStats | null>(null);

    onMount(async () => {
        try {
            const [usersRes, statsRes] = await Promise.all([
                axios.get('/api/users'),
                axios.get('/api/dashboard/stats')
            ]);
            setUsers(usersRes.data || []);
            setStats(statsRes.data);
        } catch (err) {
            console.error(err);
        }
    });

    return (
        <div class="space-y-6 animate-in">
            {/* Greeting */}
            <header class="pt-2">
                <h2 class="text-2xl font-bold text-slate-800">{t('dashboard')}</h2>
                <p class="text-slate-500 font-medium text-sm mt-1">{t('welcomeBack')}</p>
            </header>

            {/* Quick Actions (Big Buttons) */}
            <div class="grid grid-cols-2 gap-4">
                <A href="/daily-entry" class="card bg-blue-50 border-blue-200 flex flex-col items-center justify-center p-6 gap-3 active:bg-blue-100">
                    <div class="bg-blue-600 text-white p-3 rounded-full shadow-md">
                        <PlusCircle size={32} />
                    </div>
                    <span class="font-bold text-blue-900 text-center leading-tight">
                        {t('dailyEntry')}
                    </span>
                </A>

                <A href="/expenses" class="card bg-orange-50 border-orange-200 flex flex-col items-center justify-center p-6 gap-3 active:bg-orange-100">
                    <div class="bg-orange-600 text-white p-3 rounded-full shadow-md">
                        <Wallet size={32} />
                    </div>
                    <span class="font-bold text-orange-900 text-center leading-tight">
                        {t('expenses')}
                    </span>
                </A>
            </div>

            {/* Summary Cards */}
            <div class="grid grid-cols-2 gap-4">
                <div class="card bg-green-50 border-green-200">
                    <div class="flex items-center gap-2 mb-2 text-green-700">
                        <TrendingUp size={20} />
                        <span class="font-bold text-sm">{t('monthlyRevenue')}</span>
                    </div>
                    <p class="text-2xl font-black text-green-900">
                        ₹{stats()?.monthly_revenue.toLocaleString('en-IN') || '0'}
                    </p>
                </div>

                <div class="card bg-red-50 border-red-200">
                    <div class="flex items-center gap-2 mb-2 text-red-700">
                        <TrendingDown size={20} />
                        <span class="font-bold text-sm">{t('monthlyExpenses')}</span>
                    </div>
                    <p class="text-2xl font-black text-red-900">
                        ₹{stats()?.monthly_expenses.toLocaleString('en-IN') || '0'}
                    </p>
                </div>
            </div>

            <div class="card bg-indigo-50 border-indigo-200 flex items-center justify-between">
                <div>
                    <span class="font-bold text-sm text-indigo-700 flex items-center gap-2 mb-1">
                        <Users size={18} /> {t('activeCustomers')}
                    </span>
                    <p class="text-3xl font-black text-indigo-900">{stats()?.active_customers.toString() || '0'}</p>
                </div>
                <A href="/customers" class="btn btn-primary !h-10 !text-sm !w-auto">
                    {t('viewAll')}
                </A>
            </div>

            {/* Recent Customers as Cards */}
            <div class="pt-4 border-t border-slate-200">
                <h3 class="text-lg font-bold text-slate-800 mb-4">{t('recentCustomers')}</h3>
                <div class="space-y-3">
                    <For each={users().slice(0, 5)}>
                        {(user) => (
                            <div class="card p-4 flex items-center justify-between bg-white hover:border-blue-300">
                                <div class="flex items-center gap-3">
                                    <div class="w-10 h-10 bg-slate-100 rounded-full flex items-center justify-center text-slate-500">
                                        <UserIcon size={20} />
                                    </div>
                                    <div>
                                        <p class="font-bold text-slate-900">{user.name}</p>
                                        <p class="text-xs text-slate-500 font-medium">Room {user.room_no}</p>
                                    </div>
                                </div>
                                <div class="text-right">
                                    <p class={`font-black text-lg ${user.balance < 0 ? 'text-red-600' : 'text-green-600'}`}>
                                        ₹{Math.abs(user.balance).toFixed(0)}
                                        {user.balance < 0 ? ' Due' : ' Adv'}
                                    </p>
                                    <span class={`text-[10px] font-bold px-2 py-0.5 rounded-full uppercase tracking-wider ${
                                        user.plan === 'monthly' ? 'bg-blue-100 text-blue-700' : 'bg-orange-100 text-orange-700'
                                    }`}>
                                        {user.plan}
                                    </span>
                                </div>
                            </div>
                        )}
                    </For>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;
