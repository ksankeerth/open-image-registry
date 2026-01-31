/* eslint-disable @typescript-eslint/no-explicit-any */
import { DATE_FIELDS } from '../client/consts';

export const parseDates = <T extends Record<string, any>>(obj: T): T => {
  if (!obj || typeof obj !== 'object') {
    return obj;
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => parseDates(item)) as unknown as T;
  }

  const parsed: any = { ...obj };

  for (const key in parsed) {
    if (Object.prototype.hasOwnProperty.call(parsed, key)) {
      const value = parsed[key];

      if (DATE_FIELDS.includes(key) && value) {
        parsed[key] = new Date(value);
      } else if (value && typeof value === 'object') {
        parsed[key] = parseDates(value);
      }
    }
  }

  return parsed;
};
