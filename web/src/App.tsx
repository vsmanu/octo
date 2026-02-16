import { BrowserRouter, Routes, Route } from "react-router-dom";
import { Layout } from "./components/layout/Layout";
import { Dashboard } from "./pages/Dashboard";
import { Endpoints } from "./pages/Endpoints";
import { EndpointDetails } from "./pages/EndpointDetails";
import { Config } from "./pages/Config";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="endpoints" element={<Endpoints />} />
          <Route path="endpoints/:id" element={<EndpointDetails />} />
          <Route path="config" element={<Config />} />
          <Route path="settings" element={<div>Settings Page (Coming Soon)</div>} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
