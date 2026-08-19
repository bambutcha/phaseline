import type { Metadata, Viewport } from "next";
import { IBM_Plex_Sans, Orbitron } from "next/font/google";
import { Providers } from "./providers";
import "./globals.css";

const display = Orbitron({
  subsets: ["latin"],
  variable: "--font-display",
  weight: ["500", "700"],
});

const sans = IBM_Plex_Sans({
  subsets: ["latin", "cyrillic"],
  variable: "--font-sans",
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  title: "PHASELINE",
  description: "Лунная смена: два ровера, тень-терминатор, цель — 100 колонии.",
  icons: { icon: "/favicon.svg" },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  viewportFit: "cover",
  themeColor: "#05070a",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru" className={`${display.variable} ${sans.variable}`}>
      <body className="h-dvh overflow-hidden bg-ink font-sans antialiased">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
