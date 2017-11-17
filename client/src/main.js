import Vue from 'vue'
import VueRouter from 'vue-router'
import vClickOutside from 'v-click-outside'

// Plugins
import GlobalComponents from './gloablComponents'
import Notifications from './components/UIComponents/NotificationPlugin'
import SideBar from './components/UIComponents/SidebarPlugin'
import App from './App'

// router setup
import routes from './routes/routes'

// library imports
import Chartist from 'chartist'
import 'bootstrap/dist/css/bootstrap.css'
import './assets/sass/paper-dashboard.scss'
import 'es6-promise/auto'

// plugin setup
Vue.use(VueRouter)
Vue.use(GlobalComponents)
Vue.use(vClickOutside)
Vue.use(Notifications)
Vue.use(SideBar)

// configure router
const router = new VueRouter({
  routes, // short for routes: routes
  linkActiveClass: 'active'
})

// global library setup
Object.defineProperty(Vue.prototype, '$Chartist', {
  get () {
    return this.$root.Chartist
  }
})

// /* eslint-disable no-new */
// new Vue({
//   el: '#app',
//   render: h => h(App),
//   router,
//   data: {
//     Chartist: Chartist
//   }
// })

if (window.astilectron == null) {
  /* eslint-disable no-new */
  new Vue({
    el: '#app',
    render: h => h(App),
    router,
    data: {
      Chartist: Chartist
    }
  })
}

var index = {
  init: function () {
    // Wait for astilectron to be ready
    document.addEventListener('astilectron-ready', function () {
      // Listen
      index.listen()

      // Refresh list
      // index.refreshList()

      /* eslint-disable no-new */
      new Vue({
        el: '#app',
        render: h => h(App),
        router,
        data: {
          Chartist: Chartist
        }
      })

      // astilectron.send('OUESH')
    })
  },
  listen: function () {
    astilectron.listen(function (message) {
      // document.body.innerHTML = message.payload
      // switch (message.name) {
      //   case 'set.style':
      //     index.listenSetStyle(message)

      //     break
      // }
    })
  },
  listenSetStyle: function (message) {
    document.body.className = message.payload
  },
  refreshList: function () {
    astilectron.send({name: 'get.list'}, function (message) {
      if (message.payload.length === 0) {
        return
      }

      let c = `<ul>`

      for (let i = 0; i < message.payload.length; i++) {
        c += `<li class="` + message.payload[i].type + `">` + message.payload[i].name + `</li>`
      }

      c += `</ul>`
      document.getElementById('list').innerHTML = c
    })
  }
}

index.init()
