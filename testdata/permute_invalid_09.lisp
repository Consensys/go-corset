;;error:3:34-39:sorted columns must come first
(defcolumns (X :i16) (Y :i16) (Z :i16))
(defpermutation (A B C) ((+ X) Y (+ Z)))
