## Workout Service
 ### Data it owns

 1. Workout table -> a table containing all possible workouts
 2. Exercise table ->  table containing execises in a certain workout 
 3. Media table ->   table with metadata on images/videos and links to where they are actually stored 
 4. Workout-Exercise -> a table linking workout and the specific  exercises in that workout


 ### API DESIGN

 Method	    Endpoint	        Description
1. POST	    /api/workouts	    Creates a new workout definition.

2. GET	    /api/workouts	    Lists all available workouts (with filtering/pagination).

3. GET	     /api/workouts/{workoutId}	Gets the full details of a workout, including its ordered list of exercises.

4. POST	    /api/exercises	    Creates a new exercise definition.

5. GET	    /api/exercises	    Lists all available exercises.

6. POST	    /api/media/presigned-url	   (Crucial) The endpoint for Step 1 of the media upload flow. Takes { parent_id, parent_type, file_name, mime_type } and returns a pre-signed URL.

7. POST	    /api/media/upload-complete	The endpoint for Step 4 of the media upload flow, to save the metadata.