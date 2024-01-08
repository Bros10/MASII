# Data extracted from the participants screen recordings 

Twelve participants completed their assigned tasks, submitting the Task Sheet, Questionnare sheet and an `.mp4` file containing a screen-recording (no sound nor webcam) that showcased them completing the task.md sheet. Half of the participants (Control Group) did not use the tool whilst doing the task and was allowed to use any other tools. Whilst the other half (Expiermental Group) used the tool during the task. 

This document will be used to extract the timings of when each user completed each task. 

Timings will be based of when the recording begins (mins:seconds, 00:00) unless the recording showcases that an issue has occured with the web application or installing of the tool. 

**Actual Start time**
Timings below showcase when the "task" began. Thefore all noted timings were subtracted from this start time. 

User1 - 00:00
User2 - 00:00
User3 - 00:00
User4 - 02:47
User5 - 00:31
User6 - 00:00
User7 - 0:00
User8 - 01:58
User9 - 05:01
User10 - 0:48
User11 - 00:00
User12 - 00:00


## Tasks

### Task 1 - Find Hidden Page
User1 - 0:57
User2 - 1:23
User3 - 1:16
User4 - 4:25
User5 - 1:33
User6 - 1:31
User7 - 1:21
User8 - 3:52
User9 - 7:13
User10 - 1:55
User11 - 0:49
User12 - 1:12

### Task 2 - Locate email page

User1 - 1:06
User2 - 1:31
User3 - 1:23
User4 - 4:35
User5 - 1:39
User6 - 1:40
User7 - 1:29
User8 - 3:59
User9 - 7:22
User10 - 2:01
User11 - 0:57
User12 - 1:19

### Task 3 - Identify the fastest path to the passwords page

User1 - 3:28
User2 - 3:11
User3 - 2:35
User4 - 6:07
User5 - 2:59
User6 - 3:19
User7 - 2:36
User8 - 5:29
User9 - 8:56
User10 - 3:00
User11 - 2:04
User12 - 2:15

### Task 4 - Identify Security Issues

User1 - XSS: 4:28, SQLi: 6:25
User2 - XSS: 3:59, SQLi: 5:56
User3 - XSS: 3:35, SQLi: 4:52
User4 - XSS: 6:25, SQLi: 8:47 
User5 - XSS: 3:42, SQLi: 5:13
User6 - XSS: 4:08, SQLi: 6:04
User7 - XSS: 3:18, SQLi: 4:50
User8 - XSS: 5:36, SQLi: 7:39 
User9 - XSS: 9:26, SQLi: 13:13
User10 - XSS: 3:43, SQLi: 5:11
User11 - XSS: 2:50, SQLi: 4:16
User12 - XSS: 3:16, SQLi: 4:42

### Task 5 - Prioritise pages for further testing

User1 - 6:47
User2 - 6:23
User3 - 5:12
User4 - 9:17
User5 - 5:31
User6 - 6:29
User7 - 5:11
User8 - 8:07
User9 - 13:45
User10 - 5:30
User11 - 4:33
User12 - 5:06

## True total task timings

User1 - 6:47
User2 - 6:23
User3 - 5:12
User4 -  6:20
User5 - 5:00
User6 - 6:29
User7 - 5:11
User8 - 6:09
User9 - 8:44
User10 - 4:42
User11 - 4:33
User12 - 5:06

# Calculations

## Non-weighted

### Control group (User1, User2, User4, User6, and User8):
(6:47 + 6:23 + 6:20 + 6:29 + 6:09) / 5 = 31:08 / 5 = 6:13.6 (average time)

### Experimental group (User3, User5, User7, User10, User11, and User12):
(5:12 + 5:00 + 5:11 + 4:42 + 4:33 + 5:06) / 6 = 29:44 / 6 = 4:57.3 (average time)

### Percentage difference between Experimental Group and Control Group:
((6:13.6 - 4:57.3) / 6:13.6) * 100 = (1:16.3 / 6:13.6) * 100 ≈ 20.42%

Experimental Group completed the tasks 20.42% faster. 

## Weighted



### Control group (User1, User2, User4, User6, and User8):

Experience (in months): 72, 60, 36, 24, 12
Weighted total time (time * experience):



(407 * 72) + (383 * 60) + (390 * 36) + (389 * 24) + (389 * 12)  + (544 * 6) = 83592

Total experience (in months): 72 + 60 + 36 + 24 + 12 = 204

Weighted average time (weighted total time / total experience): 83592 / 204 ≈ 409.76 (rounded)

### Experimental group (User3, User5, User7, User10, User11, and User12):

Experience (in months): 144, 24, 48, 8, 8, 10

Weighted total time (time * experience):
(350 * 144) + (300 * 24) + (323 * 48) + (282 * 8) + (273 * 8) + (306 * 10) = 80604

Total experience (in months): 144 + 24 + 48 + 8 + 8 + 10 = 242

Weighted average time (weighted total time / total experience): 80604 / 242 ≈ 333.07(rounded)

### Percentage difference between Experimental Group and Control Group:
((409.76 - 333.07) / 409.76) * 100 = ≈ 18.72%

Using expierence in months as a weight, the tool allows users to complete the tasks approximately 18.72% faster on average.
