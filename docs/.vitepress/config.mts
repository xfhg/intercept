import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "INTERCEPT",
  description: "Policy as Code Engine",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Documentation', link: '/docs/tbd' },
      { text: 'Features', link: '/docs/sandbox' }
    ],
    head: [['link', { rel: 'icon', href: '/favicon.png' }]],
    logo: '/interceptvEE.png',
    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Manual', link: '/docs/manual' },
          { text: 'Docker QuickStart', link: '/docs/quickstart' },
          { text: 'Sandbox Playground', link: '/docs/sandbox' },
        ]
      },
      {
        text: 'Policy Features',
        items: [
          { text: 'Struct', link: '/docs/policy-struct' },
          { text: 'Enforcement Levels', link: '/docs/enforcement' },
          { text: 'Intercept Feature Flags', link: '/docs/feature-flags' }
        ]
      },
      {
        text: 'Policy Types',
        items: [
          { text: 'SCAN ', link: '/docs/policy-scan' },
          { text: 'ASSURE ', link: '/docs/policy-assure-regex' },
          { text: 'ASSURE - REGO ', link: '/docs/policy-assure-rego' },
          { text: 'ASSURE - TYPE ', link: '/docs/policy-assure-filetype' },
          { text: 'API ', link: '/docs/policy-assure-api' },
          { text: 'COLLECT ', link: '/docs/policy-collect' },
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/xfhg/intercept' }
    ],
    footer: {
      message: 'Released under the <a href="https://github.com/xfhg/intercept/blob/master/LICENSE">AGPL-3.0 License</a>',
      copyright: 'Copyright © 2018-2024 - <a href="https://github.com/xfhg">Flávio HG</a>'
    }
  },
  
})
