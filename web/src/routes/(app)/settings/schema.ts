import { z } from 'zod';

/** 1) Profile schema **/
export const profileSchema = z.object({
  full_name: z.string()
    .min(1, 'Full name is required')
    .regex(/^[a-zA-Z\s'-]+$/, 'Full name can only contain letters, spaces, hyphens, and apostrophes')
    .max(50, 'Full name must be less than 50 characters'),

  phone: z.string()
    .optional()
    .refine((val) => {
      if (!val || val === '') return true; // Allow empty
      // Remove all non-digits and check if it's a valid length
      const digitsOnly = val.replace(/\D/g, '');
      return digitsOnly.length >= 10 && digitsOnly.length <= 15;
    }, 'Phone number must be between 10-15 digits')
    .refine((val) => {
      if (!val || val === '') return true; // Allow empty
      // Check if it contains only numbers, spaces, hyphens, parentheses, and plus
      return /^[\d\s\-\(\)\+]+$/.test(val);
    }, 'Phone number can only contain numbers, spaces, hyphens, parentheses, and plus signs')
});
export type ProfileSchema = typeof profileSchema;

/** 2) Password schema **/
export const passwordSchema = z
  .object({
    current_password: z.string().min(1, 'Current password is required'),
    new_password: z.string().min(6, 'Must be at least 6 characters'),
    confirm_password: z.string().min(6, 'Must confirm new password'),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    path: ['confirm_password'],
    message: "Passwords must match",
  });
export type PasswordSchema = typeof passwordSchema;

/** 3) Preferences schema **/
export const preferencesSchema = z.object({
  theme: z.enum(['light', 'dark', 'system']),
  notifications: z.object({
    email_alerts: z.boolean(),
    push_alerts: z.boolean(),
    alert_frequency: z.enum(['immediate', 'hourly', 'daily']),
  }),
  dashboard: z.object({
    refresh_interval: z.number().min(10).max(300),
    default_time_range: z.enum(['15m', '1h', '6h', '24h', '7d']),
    show_system_metrics: z.boolean(),
  }),
});
export type PreferencesSchema = typeof preferencesSchema;
