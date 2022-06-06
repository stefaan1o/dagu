import React from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import DashboardLayout from "./DashboardLayout";
import Dashboard from "./pages/Dashboard";
import WorkflowDetail from "./pages/WorkflowDetails";
import WorkflowList from "./pages/WorkflowList";

type Config = {
  title: string;
  navbarColor: string;
};

type Props = {
  config: Config;
};

function App({ config }: Props) {
  return (
    <BrowserRouter>
      <DashboardLayout {...config}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="" element={<WorkflowList />} />
          <Route path="/dags/" element={<WorkflowList />} />
          <Route path="/dags/:name" element={<WorkflowDetail />} />
        </Routes>
      </DashboardLayout>
    </BrowserRouter>
  );
}

export default App;
