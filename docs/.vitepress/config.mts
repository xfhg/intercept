import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "INTERCEPT",
  description: "Policy as Code Engine",
  lastUpdated: true,
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/intercept-icon.svg' }],
  ],
  themeConfig: {
    search: {
      provider: 'local'
    },
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'Code', link: 'https://github.com/xfhg/intercept' },
      { text: 'Documentation', link: '/docs/architecture' },
      { text: 'Basics', link: '/docs/basics' },
      { text: 'Latest Release', link: 'https://github.com/xfhg/intercept/releases' }
    ],
    // head: [['link', { rel: 'icon', href: '/intercept-icon.svg' }]],
    logo: '/intercept-icon.svg',
    sidebar: [
      {
        text: 'Architecture',
        items: [
          { text: 'Workflow', link: '/docs/architecture' },
          { text: 'Basic Concepts', link: '/docs/basics' },
          { text: 'Features', link: '/docs/features' },
          
        ]
      },
      {
        text: 'Getting Started',
        items: [
          { text: 'Platform Build', link: '/docs/platform-build' },
          { text: 'Docker QuickStart', link: '/docs/docker-quickstart' },
          { text: 'Sandbox Playground', link: '/docs/sandbox' },
        ]
      },
      {
        text: 'Policy Features',
        items: [
          { text: 'Schema', link: '/docs/tbd' },
          { text: 'Enforcement Levels', link: '/docs/tbd' },
          { text: 'Patching', link: '/docs/tbd' },
          { text: 'Monitoring', link: '/docs/tbd' },
        ]
      },

      {
        text: 'Policy Types',
        items: [
          { text: 'SCAN ', link: '/docs/tbd' },
          { text: 'ASSURE ', link: '/docs/tbd' },
          { text: 'ASSURE - REGO ', link: '/docs/tbd' },
          { text: 'ASSURE - TYPE ', link: '/docs/tbd' },
          { text: 'ASSURE - API ', link: '/docs/tbd' },
          { text: 'RUNTIME ', link: '/docs/tbd' },
        ]
      },
      {
        text: 'INTERCEPT AUDIT',
        items: [
          { text: 'Compliance Reporting', link: '/docs/tbd' },
          { text: 'Feature Flags', link: '/docs/tbd' },
        ]
      },
      {
        text: 'INTERCEPT OBSERVE',
        items: [
          { text: 'Setup', link: '/docs/tbd' },
          { text: 'Runtime Daemon', link: '/docs/tbd' },
          { text: 'Integration Webhooks', link: '/docs/tbd' }
        ]
      },
      {
        text: 'Use Cases',
        items: [
          { text: 'Overview', link: '/docs/use-cases' },
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/xfhg/intercept' }
    ],
    footer: {
      message: 'Released under the <a href="https://github.com/xfhg/intercept/blob/master/LICENSE.md">EUPL-1.2 License</a>',
      copyright: 'Copyright © 2018-202X - <a href="https://github.com/xfhg">Flávio HG</a>'
    }
  },
  
})
