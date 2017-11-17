<template>
  <div>
    <!--Stats cards-->
    <div class="row">
      <div class="col-lg-3 col-sm-6" v-for="stats in statsCards">
        <stats-card>
          <div class="icon-big text-center" :class="`icon-${stats.type}`" slot="header">
            <i :class="stats.icon"></i>
          </div>
          <div class="numbers" slot="content">
            <p>{{stats.title}}</p>
            {{stats.value}}
          </div>
          <div class="stats" slot="footer">
            <i :class="stats.footerIcon"></i> {{stats.footerText}}
          </div>
        </stats-card>
      </div>
    </div>
  </div>
</template>
<script>
  import StatsCard from 'components/UIComponents/Cards/StatsCard.vue'
  import ChartCard from 'components/UIComponents/Cards/ChartCard.vue'
  export default {
    components: {
      StatsCard,
      ChartCard
    },
    /**
     * Chart data used to render stats, charts. Should be replaced with server data
     */
    data () {
      return {
        statsCards: [
          {
            type: 'success',
            icon: 'ti-money',
            title: 'Money',
            value: '0',
            footerText: '',
            footerIcon: 'ti-wallet'
          },
          {
            type: 'info',
            icon: 'ti-server',
            title: 'Blocks height',
            value: '0',
            footerText: 'Difficulty: 1',
            footerIcon: 'ti-dashboard'
          },
          {
            type: 'danger',
            icon: 'ti-bolt',
            title: 'Miner Hashrate',
            value: '0 h/s',
            footerText: 'Stopped',
            footerIcon: 'ti-reload'
          },
          {
            type: 'warning',
            icon: 'ti-pulse',
            title: 'DHT nodes',
            value: '0',
            footerText: 'Synced',
            footerIcon: 'ti-reload'
          }
        ]
      }
    }
    // created () {
    //   this.timer = setInterval(() => {
    //     astilectron.send({name: 'getInfos'}, (response) => {
    //       console.log(response)
    //       const infos = response.payload

    //       const wallets = infos.wallets
    //       this.statsCards[0].value = wallets.reduce((memo, item) => memo + item.amount, 0) + ''
    //       this.statsCards[0].footerText = wallets.length + ' wallets'

    //       this.statsCards[1].value = infos.blocksHeight
    //       // this.statsCards[1].footerText = minerInfo.running ? 'Running' : 'Stopped'

    //       const minerInfo = infos.minerInfo
    //       this.statsCards[2].value = minerInfo.hashrate + ' h/s'
    //       this.statsCards[2].footerText = minerInfo.running ? 'Running' : 'Stopped'

    //       this.statsCards[3].value = infos.nodesNb
    //       this.statsCards[3].footerText = infos.synced ? 'Synced' : 'Syncing...'
    //     })
    //   }, 1000)
    // },
    // destroyed () {
    //   this.timer.stop()
    // }
  }

</script>
<style>

  body, html {
    border: 4px;
  }

</style>
