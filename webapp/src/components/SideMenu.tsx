import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Divider } from 'primereact/divider';
import { MenuEntity } from '../types/app_types';

export type SideMenuProps = {
  menus: MenuEntity[];
  selectedMenuKey: string;
  availableHeight: number;
  // menuCollapsed: () => void;
};

const SideMenuList = (props: SideMenuProps) => {
  // Track collapsed state of menus
  const [menuState, setMenuState] = useState<{ [key: string]: boolean }>({});
  // Track thin/full mode
  const [isSlimMode, setIsSlimMode] = useState(false);

  const isMenuCollapsed = (menu: MenuEntity): boolean => {
    return (
      props.selectedMenuKey === menu.key ||
      menu.children.map((c) => c.key).includes(props.selectedMenuKey)
    );
  };

  useEffect(() => {
    const tempMenuState: { [key: string]: boolean } = {};
    props.menus.forEach((m) => {
      tempMenuState[m.key] = isMenuCollapsed(m);
    });
    setTimeout(() => {
      setMenuState(tempMenuState);
    }, 0);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.menus]);

  const navigate = useNavigate();

  const handleMenuClick = (link: string) => {
    navigate(link);
  };

  const handleMenuCollapse = (menu: string) => {
    // If in slim mode, expand to full mode when clicking a menu
    if (isSlimMode) {
      setIsSlimMode(false);
    }

    setMenuState((current) => {
      // if (!current[menu]) {
      //   props.menuCollapsed();
      // }
      return {
        ...current,
        [menu]: !current[menu],
      };
    });
  };

  const toggleSlimMode = () => {
    setIsSlimMode((prev) => !prev);
  };

  return (
    <>
      {/* Sidebar Container */}
      <div
        className="flex-grow-0 h-full flex flex-column relative transition-all transition-duration-300"
        style={{
          minWidth: isSlimMode ? '60px' : '20%',
          width: isSlimMode ? '60px' : '20%',
          maxHeight: props.availableHeight,
        }}
      >
        {/* Scrollable Menu Section */}
        <div
          className="flex flex-column flex-1 pt-4"
          style={{
            overflowY: 'auto',
            overflowX: 'hidden',
            scrollbarWidth: 'thin',
          }}
        >
          <button
            className="absolute border-none bg-white cursor-pointer hover:surface-100 border-circle transition-all transition-duration-200"
            onClick={toggleSlimMode}
            title={isSlimMode ? 'Expand sidebar' : 'Collapse sidebar'}
            style={{
              zIndex: 1000,
              right: isSlimMode ? '-10px' : '-10px',
              top: '8px',
              width: '24px',
              height: '24px',
              padding: '4px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
              backgroundColor: 'var(--surface-0)',
              border: '1px solid var(--surface-border)',
            }}
          >
            <span
              className={`pi ${isSlimMode ? 'pi-angle-right' : 'pi-angle-left'} text-600`}
              style={{ fontSize: '0.75rem' }}
            />
          </button>

          {/* Menu Items */}
          <div className="flex flex-column w-full gap-2 relative">
            {/* Toggle Button - Positioned absolutely on first menu */}

            {props.menus.map((menu) => (
              <div
                key={menu.key}
                className="font-medium text-sm flex flex-column"
                style={{ paddingLeft: isSlimMode ? '0' : '1rem' }}
              >
                <div className="flex flex-column">
                  {/* Menu Header */}
                  <div
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        handleMenuCollapse(menu.key);
                      }
                    }}
                    className={
                      'flex flex-row cursor-pointer hover:surface-50 p-2 border-round transition-colors transition-duration-200' +
                      (isSlimMode ? 'justify-content-center' : 'justify-content-between')
                    }
                    onClick={() => handleMenuCollapse(menu.key)}
                    title={isSlimMode ? menu.name : ''}
                  >
                    <span className="text-base font-medium text-800 flex align-items-center">
                      <span className={'pi ' + menu.icon_class} />
                      {!isSlimMode && (
                        <>
                          &nbsp;&nbsp;&nbsp;
                          {menu.name}
                        </>
                      )}
                    </span>
                    {!isSlimMode && (
                      <>
                        {menuState[menu.key] && <span className="pi pi-angle-up" />}
                        {!menuState[menu.key] && <span className="pi pi-angle-down" />}
                      </>
                    )}
                  </div>

                  {/* Submenu Items */}
                  {!isSlimMode &&
                    menuState[menu.key] &&
                    menu.children.map((submenu) => (
                      <div
                        role="button"
                        tabIndex={0}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' || e.key === ' ') {
                            handleMenuClick(submenu.nav_link);
                          }
                        }}
                        key={submenu.key}
                        className={
                          submenu.key === props.selectedMenuKey
                            ? 'grid grid-nogutter  cursor-pointer surface-50 border-round transition-colors transition-duration-200'
                            : 'grid grid-nogutter cursor-pointer hover:surface-50 border-round transition-colors transition-duration-200'
                        }
                        onClick={() => handleMenuClick(submenu.nav_link)}
                      >
                        <div className="col-1"></div>
                        <div className="col-10 pb-3 pt-3">{submenu.name}</div>
                        {submenu.key === props.selectedMenuKey && (
                          <div className="col-1 flex justify-content-end">
                            <span
                              className="bg-teal-400 h-full"
                              style={{ minWidth: '2px', height: '100%' }}
                            >
                              &nbsp;&nbsp;
                            </span>
                          </div>
                        )}
                      </div>
                    ))}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Fixed Footer Section */}
        <div
          className="flex justify-content-between align-items-center p-2 text-xs border-top-1 surface-border transition-all transition-duration-300 flex-shrink-0 bg-white"
          style={{
            height: '2rem',
            minHeight: '2rem',
          }}
        >
          {!isSlimMode ? (
            <>
              <div className="flex align-items-center gap-2">
                <span className="pi pi-github"></span>
                <span>Open Image Registry</span>
              </div>
              <div className="flex align-items-center">
                <span>v1.0.0</span>
              </div>
            </>
          ) : (
            <div
              role="button"
              tabIndex={0}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  window.open('https://github.com/ksankeerth/open-image-registry', '_blank');
                }
              }}
              className="flex align-items-center justify-content-center w-full cursor-pointer hover:surface-100 border-round"
              title="Open Image Registry v1.0.0"
              onClick={() => {
                window.open('https://github.com/ksankeerth/open-image-registry', '_blank');
              }}
            >
              <span className="text-xs">v1.0.0</span>
            </div>
          )}
        </div>
      </div>

      {/* Divider */}
      <Divider layout="vertical" className="p-0 m-0 surface-border" />
    </>
  );
};

export default SideMenuList;
