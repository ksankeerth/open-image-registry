import React from 'react';
import { createBrowserRouter } from 'react-router-dom';
import RootLayout from './components/RootLayout';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import NewAccountSetupPage from './pages/NewAccountSetupPage';
import ManagmentConsole from './pages/mgmt-console';
import RegistryConsole from './pages/reg-console';
import NamespaceMgmtPage from './pages/mgmt-console/ns/NamespaceMgmtPage';
import NamespaceViewPage from './pages/mgmt-console/NamespaceViewPage';
import UserAdministrationPage from './pages/mgmt-console/user/UserAdministrationPage';

const AppRouter = createBrowserRouter([
  {
    path: '/',
    children: [
      {
        path: '/login',
        element: <LoginPage />,
      },
      {
        path: '/account-setup/:uuid',
        element: <NewAccountSetupPage />,
      },
      {
        path: '/',
        Component: RootLayout,
        children: [
          {
            path: '',
            element: <HomePage />,
            index: true,
          },
          {
            path: '/console/management',
            element: <ManagmentConsole />,
            children: [
              {
                path: '/console/management/users',
                element: <UserAdministrationPage />,
              },
              {
                path: '/console/management/namespaces',
                element: <NamespaceMgmtPage />,
                children: [
                  {
                    path: '/console/management/namespaces/:id',
                    element: <NamespaceViewPage />,
                  },
                ],
              },
            ],
          },
          {
            path: '/console/registry',
            element: <RegistryConsole />,
            children: [],
          },
        ],
      },
    ],
  },
]);
export default AppRouter;
