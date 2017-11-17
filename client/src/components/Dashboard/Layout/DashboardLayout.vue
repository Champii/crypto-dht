<template>
  <div class="wrapper">
    <side-bar type="sidebar" :sidebar-links="$sidebar.sidebarLinks">

    </side-bar>
    <notifications>

    </notifications>
    <div class="main-panel">
      <top-navbar></top-navbar>

      <dashboard-content @click.native="toggleSidebar">

      </dashboard-content>

      <content-footer></content-footer>
    </div>
  </div>
</template>
<style lang="scss">

</style>
<script>
  import TopNavbar from './TopNavbar.vue'
  import ContentFooter from './ContentFooter.vue'
  import DashboardContent from './Content.vue'
  export default {
    components: {
      TopNavbar,
      ContentFooter,
      DashboardContent
    },
    methods: {
      toggleSidebar () {
        if (this.$sidebar.showSidebar) {
          this.$sidebar.displaySidebar(false)
        }
      }
    },
    created () {
      this.timer = setInterval(() => {
        astilectron.send({name: 'getInfos'}, (response) => {
          console.log(response)
          const infos = response.payload

          const wallets = infos.wallets
          this.statsCards[0].value = wallets.reduce((memo, item) => memo + item.amount, 0) + ''
          this.statsCards[0].footerText = wallets.length + ' wallets'

          this.statsCards[1].value = infos.blocksHeight
          // this.statsCards[1].footerText = minerInfo.running ? 'Running' : 'Stopped'

          const minerInfo = infos.minerInfo
          this.statsCards[2].value = minerInfo.hashrate + ' h/s'
          this.statsCards[2].footerText = minerInfo.running ? 'Running' : 'Stopped'

          this.statsCards[3].value = infos.nodesNb
          this.statsCards[3].footerText = infos.synced ? 'Synced' : 'Syncing...'
        })
      }, 1000)
    },
    destroyed () {
      this.timer.stop()
    }

  }

</script>
