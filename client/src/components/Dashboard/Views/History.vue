<template>
  <div class="content">
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
          tableName2: '',
          subTitle: '',
          columns: ['Amount', 'Timestamp', 'Address'],
          history: []
        }
      }
    },
    methods: {
      getInfos: function () {
        astilectron.send({name: 'getInfos'}, (response) => {
          const infos = response.payload

          this.table.history = infos.ownWaitingTx.map(item => {
            item.amount = item.amount / 100
            item.amount = '(' + item.amount.toFixed(2) + ')'
            return item
          }).reverse()

          this.table.history = this.table.history.concat(infos.history.map(item => {
            item.amount = item.amount / 100
            item.amount = item.amount.toFixed(2)
            return item
          }).reverse())
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
