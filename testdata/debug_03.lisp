(defcolumns (ST :i16) (X :i128))

(defconstraint c1 ()
  (if (!= 0 ST)
      (begin
       (debug (== X 2)))))

(defconstraint c2 (:guard ST)
  (begin
   (debug (== X 2))))
