(defpurefun ((vanishes! :@loob) x) x)

(defcolumns ST A B)
(defconstraint spills ()
  (vanishes!
   (* ST A (~ (* (shift A 3) (shift B 2))))))
