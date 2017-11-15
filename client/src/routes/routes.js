import DashboardLayout from '../components/Dashboard/Layout/DashboardLayout.vue'
// GeneralViews
import NotFound from '../components/GeneralViews/NotFoundPage.vue'

// Admin pages
import Overview from 'src/components/Dashboard/Views/Overview.vue'
import UserProfile from 'src/components/Dashboard/Views/UserProfile.vue'
import Notifications from 'src/components/Dashboard/Views/Notifications.vue'
import Icons from 'src/components/Dashboard/Views/Icons.vue'
import Maps from 'src/components/Dashboard/Views/Maps.vue'
import Typography from 'src/components/Dashboard/Views/Typography.vue'
import TableList from 'src/components/Dashboard/Views/TableList.vue'
import Wallets from 'src/components/Dashboard/Views/Wallets.vue'
import History from 'src/components/Dashboard/Views/History.vue'
import Mining from 'src/components/Dashboard/Views/Mining.vue'
import Dht from 'src/components/Dashboard/Views/Dht.vue'

const routes = [
  {
    path: '/',
    component: DashboardLayout,
    redirect: '/admin/dashboard'
  },
  {
    path: '/admin',
    component: DashboardLayout,
    redirect: '/admin/stats',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: Overview
      },
      {
        path: 'wallets',
        name: 'wallets',
        component: Wallets
      },
      {
        path: 'history',
        name: 'history',
        component: History
      },
      {
        path: 'mining',
        name: 'mining',
        component: Mining
      },
      {
        path: 'dht',
        name: 'dht',
        component: Dht
      }
      // {
      //   path: 'typography',
      //   name: 'typography',
      //   component: Typography
      // },
      // {
      //   path: 'table-list',
      //   name: 'table-list',
      //   component: TableList
      // }
    ]
  },
  { path: '*', component: NotFound }
]

/**
 * Asynchronously load view (Webpack Lazy loading compatible)
 * The specified component must be inside the Views folder
 * @param  {string} name  the filename (basename) of the view to load.
function view(name) {
   var res= require('../components/Dashboard/Views/' + name + '.vue');
   return res;
};**/

export default routes
