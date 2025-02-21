;;error:3:16-21:expected positive sort
(defcolumns (X :i16@prove))
(defsorted s1 ((- X)))
