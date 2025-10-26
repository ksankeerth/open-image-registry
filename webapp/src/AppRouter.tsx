import React from "react";
import { createBrowserRouter } from "react-router-dom";
import RootLayout from "./components/RootLayout";
import HomePage from "./pages/HomePage";
import RepositoryViewPage from "./pages/RepositoryViewPage";
import LoginPage from "./pages/LoginPage";
import NewAccountSetupPage from "./pages/NewAccountSetupPage";
import UserAdministrationComponent from "./components/UserAdministration";
import RegistryConsolePage from "./pages/RegistryConsolePage";

const AppRouter = createBrowserRouter([
  {
    path: "/",
    // Component: RootLayout,
    children: [
      {
        path: "/login",
        element: <LoginPage />,
      },
      {
        path: "/account-setup/:uuid",
        element: <NewAccountSetupPage />,
      },
      {
        path: "/",
        Component: RootLayout,
        children: [
          {
            path: "",
            element: <HomePage />,
            index: true,
          },
          {
            path: "/console",
            element: <RegistryConsolePage />,
            children: [
              {
                path: "/console/user-management/users",
                element: <UserAdministrationComponent />

              },
            ],
          },
          {
            path: "/repository/:repository_name",
            element: <RepositoryViewPage />,
          },
        ],
      },
    ],
  },
]);
export default AppRouter;