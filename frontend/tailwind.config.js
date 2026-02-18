/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/*.html",     
  ],
  darkMode: 'class',
  safelist: [
    'bg-blue-500',
    'bg-red-500',
    'bg-green-500',
    'text-green-600',
    'text-red-600',
    {
      pattern: /^bg-(blue|red|green)-500$/,
    },
    {
      pattern: /^text-(green|red)-600$/,
    },
    {
      pattern: /^hover:/,
    },
  ],
  theme: {
    extend: {
      colors: {
        background: 'var(--background)',
        foreground: 'var(--foreground)',
        muted: 'var(--muted)',
        'muted-foreground': 'var(--muted-foreground)',
        primary: 'var(--primary)',
        'primary-foreground': 'var(--primary-foreground)',
        card: 'var(--card)',
        'card-foreground': 'var(--card-foreground)',
        border: 'var(--border)',
        input: 'var(--input)',
        ring: 'var(--ring)',
      },
    },
  },
  plugins: [],
}