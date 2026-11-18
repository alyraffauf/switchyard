import { expect, test } from "bun:test";

test("a docs breadcrumb identifies the article as the current page", async () => {
  // Build the website before running this test.
  const html = Bun.file(
    new URL("../dist/docs/using/index.html", import.meta.url),
  );
  const links = [];
  const rewriter = new HTMLRewriter().on('nav[aria-label="Breadcrumb"] a', {
    element(link) {
      const href = link.getAttribute("href");
      if (href.startsWith("/")) {
        links.push([href, link.getAttribute("aria-current")]);
      }
    },
  });
  await rewriter.transform(new Response(html)).text();

  expect(links).toEqual([
    ["/", null],
    ["/docs/", null],
    ["/docs/using/", "page"],
  ]);
});
