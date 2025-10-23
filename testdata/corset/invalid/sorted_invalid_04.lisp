;;error:3:24-29:sorted columns must come first
(defcolumns (X :i16@prove) (Y :i16@prove) (Z :i16@prove))
(defsorted s1 ((↓ X) Y (↓ Z)))
