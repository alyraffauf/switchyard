import { useState, useEffect } from "react";
import icon from "/icon.svg";
import { Icon } from "./components/Icon";
import { screenshots } from "./data/screenshots";
import { features } from "./data/features";
import { heroButtons } from "./data/heroButtons";
import { showcases } from "./data/showcases";
import { installOptions } from "./data/installOptions";
import { footerLinks } from "./data/footerLinks";
import { DEV_URL } from "./data/urls";

function App() {
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);

  // Close lightbox on Escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setLightboxSrc(null);
    };
    if (lightboxSrc) {
      document.addEventListener("keydown", handleKeyDown);
      return () => document.removeEventListener("keydown", handleKeyDown);
    }
  }, [lightboxSrc]);

  const openLightbox = (src: string) => setLightboxSrc(src);

  return (
    <>
      {/* Hero */}
      <section className="min-h-screen flex items-center bg-gradient-to-br from-bg to-bg-alt relative overflow-hidden">
        {/* Background decoration */}
        <div className="absolute -top-1/2 -right-1/5 w-4/5 h-[200%] bg-[radial-gradient(ellipse,rgba(79,137,93,0.08)_0%,transparent_70%)] pointer-events-none" />

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 md:gap-16 items-center max-w-[1200px] mx-auto px-4 md:px-8 py-8 md:py-16 relative z-10">
          <div className="contents md:block text-center md:text-left">
            <div className="flex items-center gap-3 md:gap-4 justify-center md:justify-start mb-4 order-1">
              <img
                src={icon}
                alt=""
                className="w-14 h-14 md:w-[72px] md:h-[72px] shrink-0"
              />
              <h1 className="text-3xl md:text-[3.5rem] font-bold leading-tight tracking-tight bg-gradient-to-br from-text to-text-muted bg-clip-text text-transparent">
                Switchyard
              </h1>
            </div>
            <p className="text-lg md:text-xl text-text-muted font-medium mb-4 order-3">
              A rules-based browser launcher for Linux.
            </p>
            <p className="text-base md:text-lg text-text-muted leading-relaxed mb-4 max-w-[480px] mx-auto md:mx-0 order-4">
              Work links in Chrome. Personal links in Firefox.
            </p>
            <div className="inline-block px-3 py-2 md:px-4 md:py-2 bg-primary/10 text-primary rounded-full text-sm font-semibold mb-4 md:mb-8 order-5">
              100% Free & Open Source
            </div>
            <div className="flex flex-col md:flex-row gap-3 md:gap-4 items-center md:items-start justify-center md:justify-start order-6">
              {heroButtons.map((btn) => (
                <a
                  key={btn.label}
                  href={btn.href}
                  className={`inline-flex items-center gap-2.5 px-6 py-3.5 rounded-xl font-semibold text-base transition-all duration-200 w-full md:w-auto max-w-[260px] justify-center focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2 ${
                    btn.primary
                      ? "bg-primary text-white shadow-primary hover:bg-primary-light hover:-translate-y-0.5 hover:shadow-primary-hover"
                      : "bg-transparent text-text border-2 border-border hover:border-primary hover:text-primary"
                  }`}
                >
                  {btn.icon && (
                    <Icon name={btn.icon} filled className="w-5 h-5" />
                  )}
                  {btn.label}
                </a>
              ))}
            </div>
          </div>
          <div className="order-2 md:order-none">
            <img
              src={screenshots[0].src}
              alt={screenshots[0].alt}
              onClick={() => openLightbox(screenshots[0].src)}
              role="button"
              tabIndex={0}
              onKeyDown={(e) =>
                e.key === "Enter" && openLightbox(screenshots[0].src)
              }
              className="w-full max-w-[280px] md:max-w-[480px] h-auto rounded-xl md:rounded-2xl shadow-lg border border-border cursor-pointer transition-transform duration-300 hover:scale-[1.02] focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2 mx-auto"
            />
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="py-12 md:py-20 bg-bg">
        <div className="max-w-[1100px] mx-auto px-4 md:px-8">
          <h2 className="text-2xl md:text-3xl font-bold text-center mb-8 md:mb-10 tracking-tight">
            Key Features
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 md:gap-10 max-w-[400px] md:max-w-none mx-auto">
            {features.map((feature) => (
              <div key={feature.title} className="text-center p-4 md:p-8">
                <div className="w-11 h-11 md:w-12 md:h-12 mx-auto mb-4 bg-gradient-to-br from-primary to-primary-dark rounded-xl flex items-center justify-center text-white">
                  <Icon name={feature.icon} className="w-5 h-5 md:w-6 md:h-6" />
                </div>
                <h3 className="text-base md:text-lg font-bold mb-2">
                  {feature.title}
                </h3>
                <p className="text-text-muted text-sm leading-relaxed">
                  {feature.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Showcases */}
      {showcases.map((showcase, index) => (
        <section
          key={showcase.title}
          className={`py-12 md:py-20 ${index % 2 === 0 ? "bg-bg-alt" : "bg-bg"}`}
        >
          <div
            className={`grid grid-cols-1 md:grid-cols-[1.2fr_1fr] gap-6 md:gap-16 items-center max-w-[1200px] mx-auto px-4 md:px-8 ${
              index % 2 !== 0
                ? "md:[direction:rtl] md:[&>*]:[direction:ltr]"
                : ""
            }`}
          >
            {showcase.screenshot && (
              <img
                src={showcase.screenshot.src}
                alt={showcase.screenshot.alt}
                onClick={() => openLightbox(showcase.screenshot!.src)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) =>
                  e.key === "Enter" && openLightbox(showcase.screenshot!.src)
                }
                className="w-full max-w-[600px] h-auto rounded-xl md:rounded-2xl border border-border shadow-lg cursor-pointer transition-transform duration-300 hover:scale-[1.02] focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2 mx-auto order-first md:order-none"
              />
            )}
            {showcase.code && (
              <pre className="bg-bg-alt rounded-xl p-4 md:p-6 overflow-x-auto shadow-default border border-border">
                <code className="font-mono text-xs md:text-sm leading-relaxed text-text">
                  {showcase.code}
                </code>
              </pre>
            )}
            <div className="text-center md:text-left">
              <h2 className="text-xl md:text-2xl font-bold mb-4 tracking-tight">
                {showcase.title}
              </h2>
              <p className="text-text-muted text-base leading-relaxed max-w-[400px] mx-auto md:mx-0">
                {showcase.description}
              </p>
            </div>
          </div>
        </section>
      ))}

      {/* Installation */}
      <section
        className={`py-12 md:py-20 ${(showcases.length - 1) % 2 === 0 ? "bg-bg" : "bg-bg-alt"}`}
        id="install"
      >
        <div className="max-w-[1100px] mx-auto px-4 md:px-8">
          <h2 className="text-2xl md:text-3xl font-bold text-center mb-8 md:mb-10 tracking-tight">
            Get Started
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-[400px] md:max-w-[900px] mx-auto">
            {installOptions.map((option) => (
              <div
                key={option.title}
                className="bg-bg p-5 md:p-8 rounded-2xl border border-border transition-all duration-300 hover:shadow-default hover:-translate-y-0.5"
              >
                <div className="w-9 h-9 md:w-10 md:h-10 mb-4 text-primary">
                  <Icon name={option.icon} filled className="w-full h-full" />
                </div>
                <h3 className="text-base md:text-lg font-bold mb-1">
                  {option.title}
                </h3>
                <p className="text-text-muted text-sm mb-4">
                  {option.description}
                </p>
                {"command" in option ? (
                  <code className="block bg-bg-alt px-3 py-3 md:px-4 rounded-xl text-xs font-mono overflow-x-auto border border-border">
                    {option.command}
                  </code>
                ) : (
                  <a
                    href={option.link}
                    className="block bg-primary text-white px-4 py-3 rounded-xl text-sm font-semibold text-center transition-colors duration-200 hover:bg-primary-light hover:text-white focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
                  >
                    {option.linkText}
                  </a>
                )}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer
        className={`py-8 md:py-12 ${(showcases.length - 1) % 2 === 0 ? "bg-bg-alt" : "bg-bg"}`}
      >
        <div className="max-w-[1100px] mx-auto px-4 md:px-8 flex flex-col items-center gap-5">
          <div className="flex items-center gap-3 font-semibold text-lg">
            <img src={icon} alt="Switchyard" className="w-8 h-8" />
            <span>Switchyard</span>
          </div>
          <nav className="flex gap-6 md:gap-8">
            {footerLinks.map((link) => (
              <a
                key={link.label}
                href={link.href}
                className="text-text-muted text-sm hover:text-primary"
              >
                {link.label}
              </a>
            ))}
          </nav>
          <p className="text-text-light text-sm">
            Made by{" "}
            <a
              href={DEV_URL}
              className="text-text-muted hover:text-primary transition-colors"
            >
              Aly Raffauf
            </a>{" "}
            · GPL-3.0-or-later
          </p>
        </div>
      </footer>

      {/* Lightbox */}
      {lightboxSrc && (
        <div
          className="fixed inset-0 bg-black/90 flex items-center justify-center z-[1000] cursor-pointer backdrop-blur-sm"
          onClick={() => setLightboxSrc(null)}
          role="dialog"
          aria-modal="true"
          aria-label="Image preview"
        >
          <img
            src={lightboxSrc}
            alt="Screenshot preview"
            className="max-w-[90vw] md:max-w-[90vw] max-h-[85vh] md:max-h-[90vh] rounded-lg md:rounded-xl shadow-lg"
          />
        </div>
      )}
    </>
  );
}

export default App;
