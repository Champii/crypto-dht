<template>
  <div class="row">
    <div class="col-lg-5 col-md-12">
      <div class="col-lg-5 col-sm-12">
        <wallet>
          <div class="" slot="name">
            Name
          </div>
          <div class="" slot="amount">
            Amount
          </div>
          <div class="" slot="address">
            Address
          </div>
        </wallet>
      </div>
      <div class="col-lg-5 col-sm-12" v-for="wallet in wallets">
        <wallet>
          <div class="" slot="name">
            {{wallet.name}}
          </div>
          <div class="" slot="amount">
            {{wallet.amount}}
          </div>
          <div class="" slot="address">
            {{wallet.address}}
          </div>
        </wallet>
      </div>
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
        wallets: [
          {
            name: 'main.key',
            address: '1',
            amount: '0'
          }
        ]
      }
    },
    created () {
      astilectron.send({name: 'getInfos'}, (response) => {
        const infos = response.payload

        this.wallets = infos.wallets
      })
    },
    destroyed () {
      this.timer.stop()
    }

  }

</script>
<style>

</style>
