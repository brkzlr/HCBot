# HCBot
This bot was created using discordgo to serve in the Halo Completionist Discord by assigning Halo games completion roles to requesting members.

After requesting a role through commands prefixed with "+", the bot will check Xbox Live through the free OpenXBL REST API for the user's achievements and according to their progress, either assign a completion role or show what games they need to complete to obtain the role.

# TO DO
1. Change bot to use slash application commands.
- This bot was quickly ported from it's previous version written in Python 3 using Discord.py which didn't support slash commands, hence the prefixed commands message checking.
2. Change Xbox Live checking logic from OpenXBL API to the official Xbox Live API.
- After a period of inactivty, like a few hours, OpenXBL will return a 200 OK empty html response instead of the requested json data for the achievements, resulting in the first request after this period of inactivity failing.
