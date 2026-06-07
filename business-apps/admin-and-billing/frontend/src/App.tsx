import { A, useLocation } from "@solidjs/router";
import {
  LayoutDashboard,
  PlusCircle,
  Users,
  Wallet,
  Languages,
  Utensils,
  FileText,
  X,
} from "lucide-solid";
import { type ParentComponent, For, createSignal } from "solid-js";
import { useI18n } from "./i18n";

const App: ParentComponent = (props) => {
  const location = useLocation();
  const { t, setLocale, locale } = useI18n();
  const [isSidebarOpen, setIsSidebarOpen] = createSignal(false);

  // Streamlined Navigation for Mobile First
  const navItems = () => [
    { href: "/", icon: LayoutDashboard, label: t("dashboard") },
    { href: "/daily-entry", icon: PlusCircle, label: t("dailyEntry") },
    // Menu and Billing moved to slide-out sidebar
    { href: "/expenses", icon: Wallet, label: t("expenses") },
    { href: "/customers", icon: Users, label: t("customers") },
  ];

  const toggleLanguage = () => {
    setLocale(locale() === "en" ? "bn" : "en");
  };

  const openSidebar = () => setIsSidebarOpen(true);
  const closeSidebar = () => setIsSidebarOpen(false);

  return (
    <div class="app-container flex flex-col bg-slate-50 min-h-screen relative shadow-2xl">
      {/* Top Header */}
      <header class="bg-white border-b border-slate-200 p-4 flex items-center justify-between sticky top-0 z-40 shadow-sm">
        <div class="flex items-center gap-3">
          {/* App Logo/Name (clickable to open sidebar) */}
          <button
            onClick={openSidebar}
            aria-label="Open menu"
            class="w-10 h-10 rounded-full bg-blue-600 flex items-center justify-center text-white font-bold text-xl focus:ring-2 ring-blue-200"
          >
            R
          </button>

          <div class="flex flex-col">
            <h1 class="text-xl font-bold text-slate-900 leading-tight">
              {t("adminPanel")}
            </h1>
            <span class="text-xs text-slate-500 font-medium">
              {t("adminSubtitle")}
            </span>
          </div>
        </div>

        {/* Big Language Toggle */}
        <button
          onClick={toggleLanguage}
          class="flex items-center gap-2 bg-slate-100 hover:bg-slate-200 text-slate-700 px-4 py-2 rounded-full font-bold text-sm transition-colors border border-slate-200"
        >
          <Languages size={18} class="text-blue-600" />
          <span>{locale() === "en" ? "বাংলা" : "English"}</span>
        </button>
      </header>

      {/* Slide-out Sidebar */}
      <div
        class={`fixed inset-0 z-50 pointer-events-none transition-opacity ${
          isSidebarOpen() ? "opacity-100 pointer-events-auto" : "opacity-0"
        }`}
      >
        {/* Backdrop */}
        <div
          class={`absolute inset-0 bg-black/40 transition-opacity ${isSidebarOpen() ? "opacity-100" : "opacity-0"}`}
          onClick={closeSidebar}
        />

        {/* Panel */}
        <aside
          class={`absolute left-0 top-0 bottom-0 w-72 bg-white shadow-xl transform transition-transform p-4 ${
            isSidebarOpen() ? "translate-x-0" : "-translate-x-full"
          }`}
        >
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-blue-600 flex items-center justify-center text-white font-bold text-xl">
                R
              </div>
              <div>
                <div class="font-black text-slate-900">{t("adminPanel")}</div>
                <div class="text-xs text-slate-500">{t("adminSubtitle")}</div>
              </div>
            </div>
            <button
              onClick={closeSidebar}
              aria-label="Close menu"
              class="p-2 rounded-full hover:bg-slate-100"
            >
              <X />
            </button>
          </div>

          <nav class="space-y-3">
            <A
              href="/billing"
              class="flex items-center gap-3 p-3 rounded-lg hover:bg-slate-50"
              onClick={closeSidebar}
            >
              <FileText />
              <span class="font-bold">{t("invoice")}</span>
            </A>
            <A
              href="/menu-pricing"
              class="flex items-center gap-3 p-3 rounded-lg hover:bg-slate-50"
              onClick={closeSidebar}
            >
              <Utensils />
              <span class="font-bold">{t("menu")}</span>
            </A>
          </nav>

          <div class="mt-8">
            <h4 class="text-xs text-slate-400 uppercase tracking-wider mb-2">
              {t("menuPricing")}
            </h4>
            <div class="space-y-1">
              <For each={navItems()}>
                {(item) => (
                  <A
                    href={item.href}
                    class="flex items-center gap-3 p-2 rounded-md hover:bg-slate-50"
                    onClick={closeSidebar}
                  >
                    <item.icon />
                    <span class="font-medium">{item.label}</span>
                  </A>
                )}
              </For>
            </div>
          </div>
        </aside>
      </div>

      {/* Main Content Area */}
      <main class="flex-1 p-4 pb-28 overflow-y-auto animate-in bg-slate-50">
        {props.children}
      </main>

      {/* Bottom Navigation */}
      <nav class="fixed bottom-0 w-full max-w-[600px] bg-white border-t border-slate-200 flex justify-around items-center p-2 pb-safe shadow-[0_-4px_20px_rgba(0,0,0,0.05)] z-50">
        <For each={navItems()}>
          {(item) => (
            <MobileNavItem
              href={item.href}
              icon={item.icon}
              label={item.label}
              active={location.pathname === item.href}
            />
          )}
        </For>
      </nav>
    </div>
  );
};

const MobileNavItem = (props: {
  href: string;
  icon: any;
  label: string;
  active: boolean;
}) => (
  <A
    href={props.href}
    class={`flex flex-col items-center justify-center p-2 rounded-2xl min-w-[72px] transition-all ${
      props.active ? "text-blue-600" : "text-slate-500 hover:text-slate-900"
    }`}
  >
    <div class={`p-2 rounded-xl mb-1 ${props.active ? "bg-blue-100/50" : ""}`}>
      <props.icon
        size={24}
        class={props.active ? "stroke-[2.5]" : "stroke-2"}
      />
    </div>
    <span
      class={`text-[11px] font-bold tracking-wide leading-none ${props.active ? "text-blue-700" : ""}`}
    >
      {props.label}
    </span>
  </A>
);

export default App;
