import { Sidebar } from "primereact/sidebar";
import React, { useEffect, useLayoutEffect, useRef, useState } from "react";
import LogoComponent from "../../components/LogoComponent";
import { Button } from "primereact/button";
import { Divider } from "primereact/divider";
import { BreadCrumb } from "primereact/breadcrumb";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import SideMenuList from "../../components/SideMenu";
import { CONSOLE_BREADCRUMP_MAP } from "../../config/breadcrump";
import { CONSOLE_MENUS, MENU_KEY_NAV_LINKS } from "../../config/menus";


const RegistryConsolePage = () => {

  const location = useLocation();
  const navigate = useNavigate();

  const [visible, setVisible] = useState<boolean>(false);

  const [breadcrumpState, setBreadcrumpState] = useState<{ label: string }[]>([])

  const [selectedMenuKey, setSelectedMenuKey] = useState<string>(CONSOLE_MENUS[0].children[0].key);

  useEffect(() => {
    if (location.pathname.startsWith("/console")) {
      setVisible(true)

      // this is to redirect to default menu users
      if (location.pathname == "/console") {
        navigate("/console/user-management/users", {
          replace: true,
          viewTransition: true
        });
      }
    } else {
        setVisible(false);
    }
  }, [location.pathname]);

  
  useEffect(() => {
    if (location.pathname == "/console") {
      navigate("/console/user-management/users", {
        replace: true,
        viewTransition: true
      });
    }
  }, [location.pathname])

  useEffect(() => {
    const currentPath = location.pathname;

    // Sort by longest route first so deeper paths match correctly
    const matched = MENU_KEY_NAV_LINKS
      .filter(r => currentPath.startsWith(r.nav_link))
      .sort((a, b) => b.nav_link.length - a.nav_link.length)[0];

    if (matched) {
      setSelectedMenuKey(matched.key)
    }

  }, [location.pathname]);

  useEffect(() => {
    const segments = location.pathname.split("/").filter(Boolean);

    let accumulatedPath = "";
    const crumbs: { label: string; path: string }[] = [];

    segments.forEach((segment) => {
      accumulatedPath += `/${segment}`;
      const label = CONSOLE_BREADCRUMP_MAP[accumulatedPath];
      if (label) {
        crumbs.push({ label, path: accumulatedPath });
      }
    });

    setBreadcrumpState(crumbs);
  }, [location.pathname]);

  const handleClickHome = () => {
    navigate("/", {
      viewTransition: true
    });
    setTimeout(() => {
      setVisible(false);
    }, 150)
  }

  const refDivider = useRef<Divider>(null);
  const refBottomWidget = useRef<HTMLDivElement>(null);

  const [scrollHeight, setScrollHeight] = useState<number>(0);

  const calculateHeight = () => {
    if (refDivider.current) {
      if (!refDivider.current.getElement()) {
        return;
      }
      if (!refBottomWidget.current) {
        return;
      }

      const bottomRect = refBottomWidget.current.getBoundingClientRect();
      if (!bottomRect) {
        return;
      }

      const topRect = refDivider.current.getElement()?.getBoundingClientRect();
      if (!topRect) {
        return;
      }

      const availableHeight = (bottomRect.y + bottomRect.height) - (topRect.y + topRect.height)


      // const elementY = rect?.y; // distance from top of viewport
      // const elementHeight = rect?.height; // element height

      // // Remaining height = total viewport height - element's bottom
      // const remainingHeight = window.innerHeight - (elementY + elementHeight);
      setScrollHeight(availableHeight);
    }
  };

  const menuCollapsed = () => {
    calculateHeight();
  };

  useEffect(() => {
    // Calculate on mount
    calculateHeight();

    // // Recalculate on window resize
    // window.addEventListener("resize", calculateHeight);

    // // Cleanup
    // return () => {
    //   window.removeEventListener("resize", calculateHeight);
    // };
  }, [refBottomWidget, refDivider]);

  return (
    <Sidebar
      position="left"
      visible={visible}
      onHide={handleClickHome}
      showCloseIcon={false}
      onShow={calculateHeight}
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
      <div className="flex flex-row  p-0 m-0 h-full" style={scrollHeight ? {
        overflow: 'hidden',
        maxHeight: scrollHeight
      } : { overflow: 'hidden' }}>
        <div
          className="flex-grow-0 shadow-1 h-full flex flex-column justify-content-between pt-4"
          style={{ minWidth: "20%" }}
        >
          <div
          >
            <SideMenuList
              menus={CONSOLE_MENUS}
              selectedMenuKey={selectedMenuKey}
              menuCollapsed={menuCollapsed}
            />
          </div>
          <div className="absolute bottom-0 left-0  flex justify-content-between align-items-center p-2 text-xs border-top-1 surface-border"
            style={{
              minWidth: '20%'
            }}
          >
            <div className="flex align-items-center" ref={refBottomWidget}>
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
                overflowX: 'hidden',
                maxHeight: `${scrollHeight}px`,
              }
              : {}
          }
        >
          <div>
            <BreadCrumb
              model={breadcrumpState}
              className="surface-100 border-none text-sm p-0 pt-1 pl-2 m-0"
            />
          </div>
          <div className="w-full h-full">
            <Outlet />
          </div>
        </div>
      </div>
    </Sidebar >
  );
};

export default RegistryConsolePage;