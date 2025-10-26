import { Divider } from "primereact/divider";
import { classNames } from "primereact/utils";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";

export type MenuItem = {
  name: string;
  key: string;
  description: string;
  nav_link: string;
  icon_class?: string;
  collapsed?: boolean;
  children: MenuItem[];
};

export type SideMenuProps = {
  menus: MenuItem[];
  selectedMenuKey: string;
  menuCollapsed: () => void;
};

const SideMenuList = (props: SideMenuProps) => {
  //This object will track collapsed state of menus
  const [menuState, setMenuState] = useState<{ [key: string]: boolean }>({});

  const navigate = useNavigate();

  const handleMenuClick = (link: string) => {
    navigate(link);
  }

  const handleMenuCollapse = (menu: string) => {
    setMenuState((current) => {
      if (!current[menu]) {
        props.menuCollapsed();
      }
      return {
        ...current,
        [menu]: !current[menu],
      };
    });
  };

  return (
    <div className="flex flex-column w-full gap-2">
      {props.menus.map((menu) => (
        <div className="font-medium text-sm flex flex-column pl-4">
          <div className="flex flex-column">
            <div
              className="flex flex-row justify-content-between  cursor-pointer hover:surface-50 p-2"
              onClick={() => {
                handleMenuCollapse(menu.key);
              }}
            >
              <span className="font-semibold text-sm flex align-items-center">
                <span className={"pi " + menu.icon_class}></span>
                &nbsp; &nbsp;&nbsp;
                {menu.name}
              </span>
              {(menuState[menu.key] || menu.collapsed) && (
                <span className="pi pi-angle-up" />
              )}
              {!(menuState[menu.key] || menu.collapsed) && (
                <span className="pi pi-angle-down" />
              )}
            </div>
            {(menuState[menu.key] || menu.collapsed) &&
              menu.children.map((submenu) => (
                <div
                  className={
                    submenu.key == props.selectedMenuKey
                      ? "grid grid-nogutter pt-2 pb-2 cursor-pointer surface-100"
                      : "grid grid-nogutter pt-2 pb-2 cursor-pointer hover:surface-50"
                  }
                  onClick={() => handleMenuClick(submenu.nav_link)}
                >
                  <div className="col-1"> </div>
                  <div className="col-10">{submenu.name}</div>
                  {submenu.key == props.selectedMenuKey && (
                    <div className="col-1">
                      <span className=" text-sm pi pi pi-sort-down-fill text-teal-600 rotate-90"></span>
                    </div>
                  )}
                </div>
              ))}
          </div>
        </div>
      ))}
    </div>
  );
};

export default SideMenuList;