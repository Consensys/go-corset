;;error:4:17-20:too few target columns
;;error:4:22-23:missing sort direction
(defcolumns (X :i16@prove))
(defpermutation (Z) (X X))
