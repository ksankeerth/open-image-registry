import { MenuEntity } from '../types/app_types';

export const flattenMenus = (menus: MenuEntity[]): MenuEntity[] => {
  const result: MenuEntity[] = [];

  const walk = (items: MenuEntity[]) => {
    for (const item of items) {
      result.push(item);
      if (item.children && item.children.length > 0) {
        walk(item.children);
      }
    }
  };

  walk(menus);
  return result;
};
