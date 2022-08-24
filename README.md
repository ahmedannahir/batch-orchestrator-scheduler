# Batch scheduling


## Description

The main functionality of this project consists of scheduling, configuring and monitoring batches.  

The user can schedule one or more multiple batches using Cron.  
The user also has additional options for subsequent batches in case of multiple, such as: independancy, meaning choosing if the batch can run if the previous batch threw an error, and previous batch input, meaning reading the output of the execution of the past batch. The latter can allow passing data between separate data and can be used as means of communication between batches.  
The user can additionally add a batch to run after an existing batch, this also allows all the aforementioned additional options of the subsequent batches.  
The user can also download the logs of a batch's executions by providing the id of the execution.  
The user can furthermore disable or enable a batch. Disabling removes the job running the batch from the scheduler but keeps its data in case the user wants to re-enable the batch.  

The following are functionalities we have yet to implement: Monitor a batch's execution progress, real-time monitoring of the output of the batch and allow adding another batch that runs in case of an error.

## Installation

1. [Download and install Go](https://go.dev/doc/install)
2. Create a .env file in the root folder following the .env.example file structure.
3. Make sure all the dependencies are installed by using the following command on the root folder :
```
go mod download
```
4. To run batches, the program runs the command :
```
bash script.sh
```
So, make sure your command terminal recognize "bash". If you already have GIT installed, I suggest adding the bin folder inside GIT's installation directory to PATH sytem variables, and testing the command before with a simple .sh file containing the following script before running the program :
```
echo Hello world!
```
5. Use the following command to run the program :
```
go run .
```
or
```
go run main.go
```

## Structure

The project's packages are the following :

* **routers** : Handles all the routes of the API and their endpoint functions.
* **controllers** : Handles all the HTTP requests and responses received and sent by the API.
* **services** : Handles all the business logic for the controllers.
* **jobs** : Handles all the interactions and functions called by the scheduler.
* **handlers** : Handles helping tasks for *services* and *jobs* related to os (uploading, unzipping..) or mail or non-straightforward database interactions.
* **database** : Initialize Gorm configuration and connection to the database.
* **entities** : Struct models mapped to database tables used for the ORM Gorm.
* **models** : Struct models used for mapping other than database tables (json mapping...)

## How it works

* To add a batch (or multiple), the application expects a json config (or configs) file that follows the structure of models/config.model.go and a zip file (or files) that contains a script.sh file on the root folder as an entry point for the batch. The batch file is uploaded and unzipped, a batch row is saved in the database containing the path of the destination folder after unzipping and the data of the config. It's then added to the scheduler with its id as its tag.

* To schedule batches, we use the scheduler of the [gocron package](https://pkg.go.dev/github.com/go-co-op/gocron). The scheduler tracks all the jobs assigned to it and makes sure they are passed to the executor when ready to be run. Each job in this case takes care of running a batch when it's time. Each job has essentially two functions :
    * **Get permission to run** :
Permission is always granted for individual batches and the first batch of a consecutive list of batches. For subsequent batches, it checks if the previous batch ran, if it threw an error and the independancy of the current batch. If the permission is granted the following function is called, otherwise a row of execution with status aborted is inserted in the table in the database.
    * **Run Batch** :
The function that runs the batch, when the permission to run of the batch is granted, makes sure of creating a logfile to store the output of the execution, updating the status of the batch to running status, insert an execution row with running status, logfile path and startime. It then runs the batch, creates an error logfile that stores the stderr output in case of an error and updates the batch status execution row accordingly.

* If a batch is set to run after the current batch, we use the function setEventListeners function offered by the gocron package. it accepts to functions that will be called one before and one after the job is run. We use the latter to simply run the next job by its tag/id, then the jobs will follow the aforementioned procedure to decide if the actual batch will run and runs it when it should.
