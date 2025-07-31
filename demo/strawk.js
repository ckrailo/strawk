const introduction = "# Welcome to strawk: Structured AWK\n# This variant of the AWK programming language does not iterate over records\n# segmented by newlines. Rather, it matches the input to the regexes below,\n# and runs the actions when the input matches the regex\n\n# Strawk was inspired by Rob Pike's paper \"Structural Regular Expressions\"\n# You can read it here: https://doc.cat-v.org/bell_labs/structural_regexps/se.pdf\n\n# This program takes the right hand smiley face and converts it\n# to a rectangle coordinates system\n\n#Feedback is always appreciated, please email armand.halbert@gmail.com\n\nBEGIN { \n  x=1\n  y=1 \n}\n\n# If strawk matches this regex (at least one space), it will expand the input\n# to the maximum matching string and then run the rule.\n/ +/ {\n  x += length($0)\n}\n\n\n/#+/ {\n  print \"rect\", x, x+length($0), y, y+1\n  x+=length($0)\n}\n\n/\\n/ {\n  x=1\n  y++ \n}\n";
const introductioninput = "                          ####################\n                       ###########################\n                    #################################         ##   #####\n    ######        ######################################       ##########\n ## ######      ##########    #############    ##########       ########\n #########     ##########      ###########      ###########    ########\n   #######    ###########      ###########      #######################\n   #######################    #############    ##############  ######\n    #########################################################     ####\n     ###   ###################################################     #####\n    ####   ###################################################       ####\n    ###    ############################################## #################\n   #############  #####################################   ##################\n   #############   ##################################     ############\n  ####       ####    ##############################      ####\n             #####     #########################         ###\n               ####          ###############           ####\n                #####                                #####\n                 ######      ##############        #####\n                   ########     #############   #######\n                      ###########  #################\n                         ######################\n                                 ###############\n                                     ############\n                                      ###########\n                                       ########";

const fsastate = cm6.createEditorStateForStrawk(introduction);
const fsaed = cm6.createEditorView(fsastate, document.getElementById("strawkeditor"));
const inputstate = cm6.createEditorState(introductioninput);
const inputed = cm6.createEditorView(inputstate, document.getElementById("inputeditor"));
const outputstate = cm6.createEditorState("");
const outputed = cm6.createEditorView(outputstate, document.getElementById("outputeditor"));

function runProgram() {
  var body = {
    "program" : fsaed.state.doc.toString(),
    "data" : inputed.state.doc.toString()
  };
  $.ajax("https://us-east1-ahalbert-clickstream.cloudfunctions.net/strawk-api", {
      method : 'POST',
      data : JSON.stringify(body),
      contentType : 'application/json',
      success: function(data) {
        var newState = cm6.createEditorState(data['output'])
        outputed.setState(newState);
      }
    }
  );
}

function changeExample(fsa, inp) {
  var newState = cm6.createEditorStateForStrawk(fsa);
  fsaed.setState(newState);
  newState = cm6.createEditorState(inp);
  inputed.setState(newState);
}
