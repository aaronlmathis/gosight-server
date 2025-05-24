// Set fallback prerender option to make sure dynamic routes work properly
export const prerender = false;
export const ssr = false;

// Page loads dynamically client-side
export function load() {
  return {};
}
