;;error:3:34-39:sorted columns must come first
(defcolumns X Y Z)
(defpermutation (A B C) ((+ X) Y (+ Z)))
