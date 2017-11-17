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
        selected: '',
        wallets: []
      }
    },
    created () {
      astilectron.send({name: 'getInfos'}, (response) => {
        const infos = response.payload
        this.wallets = infos.wallets
        this.selected = this.wallets[0]
      })
    },
    destroyed () {
      // this.timer.stop()
    }

  }

</script>
<style>

</style>
