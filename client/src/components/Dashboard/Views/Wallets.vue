<template>
  <div class="row">
    <div class="col-lg-12">
      <div class="col-lg-12">
        <select v-model="selected">
          <option v-for="wallet in wallets" v-bind:value="wallet">{{wallet.name}}</option>
        </select>
      </div>
      <wallet :item="selected">
      </wallet>
    </div>
  </div>
</template>
<script>
  import Wallet from 'components/UIComponents/Wallet/Wallet.vue'

  export default {
    components: {
      Wallet
    },
    data () {
      return {
        wallets: [],
        selected: ''
      }
    },
    methods: {
      getInfos: function () {
        astilectron.send({name: 'getInfos'}, (response) => {
          const infos = response.payload
          this.wallets = infos.wallets

          let pendingAmount = (infos.ownWaitingTx.reduce((memo, item) => memo + item.amount, 0) / 100)
          if (pendingAmount === 0) {
            pendingAmount = ''
          } else {
            pendingAmount = ' (' + ((pendingAmount > 0) ? ('+' + pendingAmount.toFixed(2)) : pendingAmount.toFixed(2)) + ')'
          }

          this.wallets = this.wallets.map(item => {
            item.amount = item.amount / 100
            item.amount = item.amount.toFixed(2) + pendingAmount
            return item
          })

          this.selected = this.wallets[0]

          this.table.history = infos.history.map(item => {
            item.amount = item.amount / 100
            item.amount = item.amount.toFixed(2)
            return item
          })

          this.table.waitingTx = infos.ownWaitingTx.map(item => {
            item.amount = item.amount / 100
            item.amount = item.amount.toFixed(2)
            return item
          })
        })
      }
    },
    created () {
      this.getInfos()
      this.timer = setInterval(() => {
        this.getInfos()
      }, 1000)
    },
    destroyed () {
      this.timer.stop()
    }

  }

</script>
<style>

</style>
