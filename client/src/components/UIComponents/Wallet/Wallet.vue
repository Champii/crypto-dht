<template>
  <div class="card">
    <div class="content">
      <div class="row">
        <div class="col-xs-2">
          <slot name="name"></slot>
        </div>
        <div class="col-xs-2">
          <slot name="amount"></slot>
        </div>
        <div class="col-xs-8">
          <slot name="address"></slot>
        </div>
      </div>
    </div>

  </div>
</template>
<script>
  export default {
    name: 'wallet'
  }

</script>
<style>

</style>
