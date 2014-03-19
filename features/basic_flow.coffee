casper.test.begin 'Basic Flow', 5, (test) ->
  casper.options.waitTimeout = 60 * 1000
  casper.start "http://localhost:8080/", ->
    @test.assertHttpStatus 200, 'UI is responding'

  casper.then ->
    @previous_experiments_count = @evaluate previousExperimentCount
    @echo("Currently #{@.previous_experiments_count} experiments in the previous experiments list")

    @fill 'form', 
      inputExperiment: "dummy"
      inputIterations: 7
      inputConcurrency: 5
    @click 'button[type=submit]'
    @waitWhileVisible ".noexperimentrunning"

  casper.then ->
    @test.assertUrlMatch ///
      /\#/experiments/.*
    ///

    @waitFor ->
      @evaluate ->
        $("#data tr").length == 7

  casper.then ->
    @test.assertElementCount "#data tr", 7, "As many rows in the data table as requested pushes"
    @test.assertElementCount "svg rect.bar", 7, "As many bars in the graph as requested pushes"
    @waitFor ->
      @evaluate experimentCountEquals, @previous_experiments_count + 1

  casper.then ->
    @capture "previous_experiments.png"
    @test.assertElementCount "#previousExperiments tr", @previous_experiments_count + 1, "Has one more previous experiments in the list"
  
  casper.run ->
    test.done()


previousExperimentCount = ->
  $("#previousExperiments tr").length

experimentCountEquals = (target) ->
  $("#previousExperiments tr").length == target
