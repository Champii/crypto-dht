import Sidebar from './SideBar.vue'

const SidebarStore = {
  showSidebar: false,
  sidebarLinks: [
    {
      name: 'Dashboard',
      icon: 'ti-panel',
      path: '/admin/dashboard'
    },
    {
      name: 'Wallets',
      icon: 'ti-wallet',
      path: '/admin/wallets'
    },
    {
      name: 'History',
      icon: 'ti-book',
      path: '/admin/history'
    },
    {
      name: 'Mining',
      icon: 'ti-bolt',
      path: '/admin/mining'
    },
    {
      name: 'DHT Infos',
      icon: 'ti-pulse',
      path: '/admin/dht'
    },
    {
      name: 'Debug',
      icon: 'ti-support',
      path: '/admin/debug'
    }
    // {
    //   name: 'Notifications',
    //   icon: 'ti-bell',
    //   path: '/admin/notifications'
    // }
  ],
  displaySidebar (value) {
    this.showSidebar = value
  }
}

const SidebarPlugin = {

  install (Vue) {
    Vue.mixin({
      data () {
        return {
          sidebarStore: SidebarStore
        }
      }
    })

    Object.defineProperty(Vue.prototype, '$sidebar', {
      get () {
        return this.$root.sidebarStore
      }
    })
    Vue.component('side-bar', Sidebar)
  }
}

export default SidebarPlugin
