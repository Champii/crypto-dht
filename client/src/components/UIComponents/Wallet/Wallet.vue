<template>
  <div class="content">
    <div class="card">
      <div class="content">
        <div class="row">
          <div class="col-lg-12">
            <div class="">
              <label>Amount:</label> {{item.amount}}
            </div>
            <div class="">
              <label>Address:</label> {{item.address}}
              <button v-on:click="copy">Copy</button>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="card">
      <div class="content">
        <div class="row">
          <div class="col-lg-12">
            <label>Amount:</label>
            <input type="text" v-model="amount"/>
          </div>
        </div>
        <div class="row">
          <div class="col-lg-12">
            <label>Address:</label>
            <input type="text" v-model="destination"/>
          </div>
        </div>
        <div class="row">
          <div class="col-lg-4">
            <button v-on:click="send">Send !</button>
          </div>
          <div class="col-lg-4" v-if="response.length > 0">
            Response: {{response}}
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
        amount: '',
        destination: '',
        response: ''
      }
    },
    methods: {
      copy: function () {
        const input = document.createElement('input')
        input.type = 'text'
        input.value = this.item.address
        this.$el.appendChild(input)
        input.select()
        document.execCommand('copy')
        input.remove()
      },
      send: function () {
        const amount = parseInt(Number(this.amount) * 100, 10)
        const dest = this.destination

        this.amount = ''
        this.destination = ''

        astilectron.send({name: 'send', payload: amount + ':' + dest}, (response) => {
          console.log('SENT', response)
          this.response = response.payload
          if (!this.response.length) {
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
