<template>
  <div class="card">
    <div class="content">
      <div class="row">
        <div class="">
          {{item.name}}
        </div>
        <div class="">
          {{item.amount}}
        </div>
        <div class="">
          {{item.address}}
        </div>
      </div>
      <div class="row">
        <div class="col-lg-12">
          Send:
          <div class="row">
            <span class="col-lg-4">
              Amount
            </span>
            <span class="col-lg-4">
              <input type="text" v-model="amount"/>
            </span>
          </div>
          <div class="row">
            <span class="col-lg-4">
              Destination
            </span>
            <span class="col-lg-4">
              <input type="text" v-model="destination"/>
            </span>
          </div>
          <div class="row">
            <span class="col-lg-4">
              <button v-on:click="send">Send !</button>
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
  export default {
    name: 'wallet',
    props: ['item'],
    data () {
      return {
        amount: 0,
        destination: ''
      }
    },
    methods: {
      send: function () {
        astilectron.send({name: 'send', payload: this.amount + ':' + this.destination}, (response) => {
          console.log('SENT', response)
          this.amount = 0
          this.destination = ''
        })
      }
    }
  }

</script>
<style>

</style>
