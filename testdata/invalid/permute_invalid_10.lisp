;;error:3:24-29:expected positive sort
(defcolumns (X :i16) (Y :i16))
(defpermutation (A B) ((- X) Y))
