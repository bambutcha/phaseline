import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#05070a",
        panel: "#121920",
        gold: "#e0c05a",
        cold: "#5aa6d4",
        danger: "#e06868",
        ok: "#6ecf8a",
        muted: "#8b96a1",
      },
      fontFamily: {
        display: ["var(--font-display)", "ui-sans-serif", "system-ui", "sans-serif"],
        sans: ["var(--font-sans)", "ui-sans-serif", "system-ui", "sans-serif"],
      },
      boxShadow: {
        gold: "0 0 28px rgba(224,192,90,0.35)",
        cold: "0 0 24px rgba(90,166,212,0.3)",
      },
      keyframes: {
        nudge: {
          "50%": { transform: "translateY(-5px)", boxShadow: "0 0 22px rgba(224,192,90,0.45)" },
        },
        pulseLine: {
          "0%, 100%": { opacity: "0.55" },
          "50%": { opacity: "1" },
        },
        rise: {
          to: { transform: "translateY(-56px)", opacity: "0" },
        },
      },
      animation: {
        nudge: "nudge 1.15s ease-in-out infinite",
        pulseLine: "pulseLine 2.4s ease-in-out infinite",
        rise: "rise 1.1s ease-out forwards",
      },
    },
  },
  plugins: [],
};

export default config;
