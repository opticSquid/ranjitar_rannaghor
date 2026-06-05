/* render from src/index.tsx */
import { render } from "solid-js/web";
import { Router, Route } from "@solidjs/router";
import { I18nProvider } from "./i18n";
import App from "./App";
import Dashboard from "./pages/Dashboard";
import DailyEntry from "./pages/DailyEntry";
import Customers from "./pages/Customers";
import Billing from "./pages/Billing";
import Expenses from "./pages/Expenses";
import Analytics from "./pages/Analytics";
import MenuPricing from "./pages/MenuPricing";
import "./index.css";

const root = document.getElementById("root");

if (import.meta.env.DEV && !(root instanceof HTMLElement)) {
  throw new Error(
    "Root element not found. Did you forget to add it to your index.html? Or is the id misspelled?",
  );
}

render(
  () => (
    <I18nProvider>
      <Router root={App}>
        <Route path="/" component={Dashboard} />
        <Route path="/customers" component={Customers} />
        <Route path="/daily-entry" component={DailyEntry} />
        <Route path="/billing" component={Billing} />
        <Route path="/expenses" component={Expenses} />
        <Route path="/analytics" component={Analytics} />
        <Route path="/menu-pricing" component={MenuPricing} />
      </Router>
    </I18nProvider>
  ),
  root!,
);
