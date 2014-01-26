casper.test.begin 'Basic Flow', 4, (test) ->
  casper.start "http://localhost:8080/ui/", ->
    @test.assertHttpStatus 200, 'UI is responding'
    @previous_experiments_count = @evaluate ->
      $("#previousExperiments tr").length
    @log("Currently #{@.previous_experiments_count} experiments in the previous experiments list")

    @fill 'form', 
      inputPushes: 2
      inputConcurrency: 5
    @click 'button[type=submit]'
    @waitWhileVisible ".noexperimentrunning"

  casper.then ->
    @test.assertUrlMatch ///
      /ui/\#/experiments/.*
    ///

    # Change this to when experiment status is 'done' once we have experiment status
    @wait 8000 

  casper.then ->
    @test.assertElementCount "#data tr", 2, "As many rows in the data table as requested pushes"
    @test.assertElementCount "svg rect", 2, "As many bars in the graph as requested pushes"

  casper.then ->
    @capture "previous_experiments.png"
    @test.assertElementCount "#previousExperiments tr", @previous_experiments_count + 1, "Has one more previous experiments in the list"
  
  casper.run ->
    test.done()
