<template>
  <div class="card">
    <div class="content">
      <div class="row">
        <div class="col-lg-12">
          <span class="">
            {{item.amount}}
          </span>
          <span class="">
            <input id="addr" type="text" :value="item.address">
            <button v-on:click="copy">Copy</button>
          </span>
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
            <span class="col-lg-4" v-if="response.length > 0">
              Response: {{response}}
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
        destination: '',
        response: ''
      }
    },
    methods: {
      copy: function () {
        const copyText = document.getElementById('addr')
        copyText.select()
        document.execCommand('copy')
      },
      send: function () {
        const amount = parseInt(Number(this.amount) * 100, 10)
        const dest = this.destination
        this.amount = 0
        this.destination = ''
        astilectron.send({name: 'send', payload: amount + ':' + dest}, (response) => {
          console.log('SENT', response)
          this.response = response.payload
          if (this.response === '') {
            this.response = 'OK'
          }
          const timer = setTimeout(() => {
            this.response = ''
          }, 7000)
        })
      }
    }
  }

</script>
<style>

</style>
