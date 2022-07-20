# Ganymede Migrate

This tool with convert Ceres files and database entries to the new Ganymede format. All of this can be done manually if you prefer to do so.

# READ THE ENTIRE DOCUMENTATION BEFORE STARTING!

### Information

* Will only work on files that are in the old format produced by Ceres. Any custom files you added will probably error out.
* Does not delete any files, it does rename them though.
* Use this tool at your own risk and on a **fresh** Ganymede installation.
* No Ceres DB entries are deleted.

### What It Does

This tool will log into Ceres and get all the VODs in the database. It will then create a new entry in Ganymede for each VOD. Once the database entry has been created, all VOD files will be renamed to match the new naming convention used by Ganymede.

### Notes

Once issue I found while running this had to do with channel names. If a streamer changed their channel/user name, you will have to manually import those VODs from that channel. An example of this is xqc. His original name was `xqcow` now it is `xqc`. I created the channel in Ganymede with `xqc` but the Ceres channel was `xqcow`. The migration tool cannot detect this and will show an error as it cannot find the channel `xqcow`. These VODs will need to be manually migrated.

### Getting Started

1. Create each channel you have in Ceres in Ganymede. This is a **required** step. The tool will **not** auto create channels.
2. Download a copy of the `docker-compose.yml` and update the Ceres host, username, password along with the Ganymede host, username and password. **Keep `SHOULD_RENAME` commented out for now. If set to true it will rename files. It is best to run a database to database migration test run first before renaming files.**
3. Update the path to your VOD folder in the `volumes` sections. This should match the folder you have in your Ceres or Ganymede compose files.
4. Run the `docker compose up` command.

At this point the migration tool should be running and adding the Ceres VOD DB entries into Ganymede. If any errors appear now is the time to fix them as once you run the container with `SHOULD_RENAME` set to true, you **cannot** run it again.

Within the `./data` folder resides a log file of the migration and any errors.

If you need to make changes and run another dry run, **delete** all VODs in the **Ganymede** database. The script will duplicate entries if VODs are not first removed due to how the ID for each vod is generated.

If you are satisfied and want to rename VOD files follow the below steps. **Once `SHOULD_RENAME` is set to true, you cannot run the migration tool again!**

1. **Delete** all VOD entries in **Ganymede**.
2. Uncomment the `SHOULD_RENAME` variable in the `docker-compose.yml` file.
3. Run the `docker compose up` command.

It is likely a few files will error out because they could not be found. Take a look at the `./data/log.log` file and manually fix any rename errors.