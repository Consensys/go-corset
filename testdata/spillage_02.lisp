(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns ST A)
(defconstraint spills ()
  (vanishes!
   (* ST (* A (~ (shift A 2))))))
