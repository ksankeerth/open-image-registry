import { Sidebar } from "primereact/sidebar";
import React, { useEffect, useRef, useState } from "react";
import LogoComponent from "../components/LogoComponent";
import { Button } from "primereact/button";
import { Divider } from "primereact/divider";
import { BreadCrumb } from "primereact/breadcrumb";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import SideMenuList, { MenuItem } from "../components/SideMenu";


const menus: MenuItem[] = [
  {
    name: "USER MANAGEMENT",
    key: "user-management",
    nav_link: "/console/user-management",
    description: "Manage users, invitations, and monitor user activity",
    icon_class: "pi-users",
    collapsed: true,
    children: [
      {
        name: "Users",
        key: "users",
        nav_link: "/console/user-management/users",
        description: "Manage user accounts, roles, and permissions",
        children: [],
      },
      {
        name: "Invitations",
        key: "invitations",
        nav_link: "/console/user-management/invitations",
        description: "Send and manage user invitations to the registry",
        children: [],
      },
      {
        name: "Activity Log",
        key: "activity-log",
        nav_link: "/console/user-management/activity-log",
        description: "View system activity, user actions, and audit logs",
        children: [],
      },
    ],
  },
  {
    name: "ACCESS MANAGEMENT",
    key: "access-management",
    nav_link: "/console/access-management",
    description: "Configure namespaces, repositories, and upstream registries",
    icon_class: "pi-shield",
    children: [
      {
        name: "Namespaces",
        key: "namespaces",
        nav_link: "/console/access-management/namespaces",
        description: "Organize and manage namespaces for your repositories",
        children: [],
      },
      {
        name: "Repositories",
        key: "repositories",
        nav_link: "/console/access-management/repositories",
        description: "Configure and manage image repositories",
        children: [],
      },
      {
        name: "Upstreams",
        key: "upstreams",
        nav_link: "/console/access-management/upstreams",
        description:
          "Configure upstream registry connections for proxy caching",
        children: [],
      },
    ],
  },
  {
    name: "INTEGRATION",
    key: "integration",
    nav_link: "/console/integration",
    description: "Configure authentication providers and external integrations",
    icon_class: "pi-ticket",
    children: [
      {
        name: "LDAP",
        key: "ldap",
        nav_link: "/console/integration/ldap",
        description: "Configure LDAP authentication and user synchronization",
        children: [],
      },
      {
        name: "SSO",
        key: "sso",
        nav_link: "/console/integration/sso",
        description: "Configure Single Sign-On with OAuth2/OIDC providers",
        children: [],
      },
    ],
  },
  {
    name: "JOBS",
    key: "jobs",
    nav_link: "/console/jobs",
    description: "Schedule and monitor automated maintenance tasks",
    icon_class: "pi-objects-column",
    children: [
      {
        name: "Cache Cleanup",
        key: "cache-cleanup",
        nav_link: "/console/jobs/cache-cleanup",
        description: "Configure and schedule cache cleanup jobs",
        children: [],
      },
      {
        name: "Storage Cleanup",
        key: "storage-cleanup",
        nav_link: "/console/jobs/storage-cleanup",
        description: "Configure and schedule storage cleanup jobs",
        children: [],
      },
      {
        name: "Manual Jobs",
        key: "manual-jobs",
        nav_link: "/console/jobs/manual",
        description: "Run and monitor manual maintenance jobs",
        children: [],
      },
    ],
  },
  {
    name: "ANALYTICS",
    key: "analytics",
    nav_link: "/console/analytics",
    description: "Monitor system performance, usage, and health metrics",
    icon_class: "pi-wave-pulse",
    children: [
      {
        name: "Storage",
        key: "storage-analytics",
        nav_link: "/console/analytics/storage",
        description: "View storage usage statistics and trends",
        children: [],
      },
      {
        name: "Cache",
        key: "cache-analytics",
        nav_link: "/console/analytics/cache",
        description: "View cache hit rates and performance metrics",
        children: [],
      },
      {
        name: "Errors",
        key: "errors-analytics",
        nav_link: "/console/analytics/errors",
        description: "Monitor and analyze system errors and failures",
        children: [],
      },
      {
        name: "Garbage Collection",
        key: "garbage-collection",
        nav_link: "/console/analytics/garbage-collection",
        description: "View garbage collection status and history",
        children: [],
      },
    ],
  },
];

const breadcrumbMap: Record<string, string> = {
  "/console": "Registry Console",

  // USER MANAGEMENT
  "/console/user-management": "User Management",
  "/console/user-management/users": "Users",
  "/console/user-management/invitations": "Invitations",
  "/console/user-management/activity-log": "Activity Log",

  // ACCESS MANAGEMENT
  "/console/access-management": "Access Management",
  "/console/access-management/namespaces": "Namespaces",
  "/console/access-management/repositories": "Repositories",
  "/console/access-management/upstreams": "Upstreams",

  // INTEGRATION
  "/console/integration": "Integration",
  "/console/integration/ldap": "LDAP",
  "/console/integration/sso": "SSO",

  // JOBS
  "/console/jobs": "Jobs",
  "/console/jobs/cache-cleanup": "Cache Cleanup",
  "/console/jobs/storage-cleanup": "Storage Cleanup",
  "/console/jobs/manual": "Manual Jobs",

  // ANALYTICS
  "/console/analytics": "Analytics",
  "/console/analytics/storage": "Storage Analytics",
  "/console/analytics/cache": "Cache Analytics",
  "/console/analytics/errors": "Error Analytics",
  "/console/analytics/garbage-collection": "Garbage Collection",
};

const RegistryConsolePage = () => {

  const location = useLocation();
  const navigate = useNavigate();

  const [visible, setVisible] = useState<boolean>(false);

  const [breadcrumpState, setBreadcrumpState] = useState<{ label: string }[]>([])

  useEffect(() => {
    if (location.pathname.startsWith("/console")) {
      setVisible(true)
    } else {
      setVisible(false);
    }
  }, [location.pathname]);

  useEffect(() => {
    if (location.pathname == "/console") {
      navigate("/console/user-management/users")
    }
  }, [location.pathname])

  useEffect(() => {
    const segments = location.pathname.split("/").filter(Boolean);

    let accumulatedPath = "";
    const crumbs: { label: string; path: string }[] = [];

    segments.forEach((segment) => {
      accumulatedPath += `/${segment}`;
      const label = breadcrumbMap[accumulatedPath];
      if (label) {
        crumbs.push({ label, path: accumulatedPath });
      }
    });

    setBreadcrumpState(crumbs);
  }, [location.pathname]);

  const handleClickHome = () => {
    navigate("/");
    setTimeout(() => {
      setVisible(false);
    }, 150)
  }

  const refDivider = useRef<Divider>(null);
  const [scrollHeight, setScrollHeight] = useState<number>(0);

  const calculateHeight = () => {
    if (refDivider.current) {
      if (!refDivider.current.getElement()) {
        return;
      }
      // Get the position and size of the element
      const rect = refDivider.current.getElement()?.getBoundingClientRect();
      if (!rect) {
        return;
      }
      const elementY = rect?.y; // distance from top of viewport
      const elementHeight = rect?.height; // element height

      // Remaining height = total viewport height - element's bottom
      const remainingHeight = window.innerHeight - (elementY + elementHeight);
      setScrollHeight(remainingHeight);
    }
  };

  const menuCollapsed = () => {
    calculateHeight();
  };

  useEffect(() => {
    // Calculate on mount
    calculateHeight();

    // Recalculate on window resize
    window.addEventListener("resize", calculateHeight);

    // Cleanup
    return () => {
      window.removeEventListener("resize", calculateHeight);
    };
  });

  return (
    <Sidebar
      position="left"
      visible={visible}
      onHide={handleClickHome}
      showCloseIcon={false}
      header={
        <div className="flex flex-column w-full pt-2 m-0">
          <div className="flex flex-row justify-content-between w-full">
            <div>
              <LogoComponent showNameInOneLine={false} />
            </div>
            <div className="flex-grow-1 flex justify-content-center align-items-center font-semibold text-lg">
              Registry Console
            </div>
            <div className="flex align-items-center pr-3">
              <Button
                className="border-round-3xl border-1"
                outlined
                size="small"
                onClick={handleClickHome}
              >
                <span className="pi pi-chevron-left"></span>
                &nbsp;&nbsp;
                <span>Home</span>
              </Button>
            </div>
          </div>
          <Divider className="p-0 m-0" ref={refDivider} />
        </div>
      }
      style={{ width: "100vw" }}
    >
      <div className="flex flex-row  p-0 m-0 h-full">
        <div
          className="flex-grow-0 shadow-1 h-full flex flex-column justify-content-between pt-4"
          style={{ width: "20%" }}
        >
          <div
            style={
              scrollHeight
                ? {
                  overflowY: "auto",
                  maxHeight: `${scrollHeight * 0.9}px`,
                }
                : {}
            }
          >
            <SideMenuList
              menus={menus}
              selectedMenuKey={menus[0].children[0].key}
              menuCollapsed={menuCollapsed}
            />
          </div>
          <div className="flex justify-content-between align-items-center p-2 text-xs border-top-1 surface-border">
            <div className="flex align-items-center">
              <span className="pi pi-github"></span>
              &nbsp;&nbsp;
              <span>Open Image Registry</span>
            </div>
            <div className="flex align-items-center">
              <span>v1.0.0</span>
            </div>
          </div>
        </div>
        <Divider className="p-0 m-0" layout="vertical" />
        <div
          className="flex-grow-1 surface-100 flex flex-column p-2"
          style={
            scrollHeight
              ? {
                overflowY: "auto",
                maxHeight: `${scrollHeight}px`,
              }
              : {}
          }
        >
          <div>
            <BreadCrumb
              model={[
                { label: "Registry Console" },
                { label: "User Management" },
                { label: "Users" },
              ]}
              className="surface-100 border-none text-sm p-0 pt-1 pl-2 m-0"
            />
          </div>
          <div className="w-full h-full">
            <Outlet />
          </div>
        </div>
      </div>
    </Sidebar>
  );
};

export default RegistryConsolePage;