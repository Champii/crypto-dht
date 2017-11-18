<template>
  <div class="content">
    <div class="row">
      <div class="col-md-12">
        <div class="card">
          <paper-table :title="table.tableName" :sub-title="table.subTitle" :data="table.waitingTx" :columns="table.columns">
          </paper-table>
        </div>
      </div>
    </div>
    <div class="row">
      <div class="col-md-12">
        <div class="card">
          <paper-table :title="table.tableName2" :sub-title="table.subTitle" :data="table.history" :columns="table.columns">
          </paper-table>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
  import PaperTable from 'components/UIComponents/PaperTable.vue'

  export default {
    components: {
      PaperTable
    },
    data () {
      return {
        table: {
          tableName: 'Waiting tx',
          tableName2: 'History tx',
          subTitle: '',
          columns: ['Amount', 'Timestamp', 'Address'],
          history: [],
          waitingTx: []
        }
      }
    },
    methods: {
      getInfos: function () {
        astilectron.send({name: 'getInfos'}, (response) => {
          const infos = response.payload

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
      }, 10000)
    },
    destroyed () {
      this.timer.stop()
    }

  }

</script>
<style>

</style>
